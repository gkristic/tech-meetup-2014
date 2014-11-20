Concurrency
===========

The [original][] code was modified to run checksums concurrently, by recursively
descending into the tree and triggering an asynchronous sum over each different
file that is found. I'll summarize the updates here.

  [original]: https://github.com/gkristic/tech-meetup-2014/tree/original

### Concurrent processing is not always better

First, please note that **this is not necessarily better** than the version we
had in terms of running time. Both rsum and fdup are dominated by I/O time;
especially disk throughput. Running concurrently, or even in parallel if you
have more than one core, does not increase your data flow. Actually, running
concurrently can make the situation even worse (i.e., make the application take
longer to finish) because you're changing what's mostly a sequential read by a
lot of random accesses. (The difference is way bigger for magnetic disks, when
compared to SSDs.)

So, keep in mind that **this is mainly for demonstration** purposes, as opposed
to showing production-ready applications. I still think it's more fun to show
useful tools than a handful of useless toy programs. These tools *will* probably
perform better, though, if the tree you're working with spans several different
physical drives. In such a scenario, concurrent execution will take advantage of
the total throughput, while a serial run will still visit files one by one, with
all but one drive at a time sitting idle. Higher quality tools should probably
consider the distinction in underlying storage, running concurrently only those
files that are detected to live in different drives. But that's left as an
exercise for the reader :)

Also note that, even though we've updated the tools to visit files concurrently,
that doesn't mean processing is done in parallel. In fact, Go uses a single OS
thread for user code unless told otherwise (at least up to Go 1.3.3, the latest
at the time of this writing), meaning that you're using a single core at a time.
You can tell Go's runtime to increase the maximum number of OS threads to use by
setting `GOMAXPROCS`, although Go will spawn more threads regardless when
blocked on system calls. That's not necessarily better (otherwise Go wouldn't
use 1 by default), but it can help.

### Enabling concurrency

As we saw during the workshop, starting a goroutine is quite easy. You just say
`go f()`, where `f()` is the function you want to run asynchronously. So the
first step was to start several routines for the arguments in the command line.
You can see the result [here][rsum:main-loop]. We used the `sync` package to
wait for all routines to finish. Otherwise, the application would exit even
before they have a chance to run. Now the same thing can be done inside the
`Walker`, in the recursive step, so that each object in a directory is dealt
with in a separate routine as well. But in this case we need to wait for a
result, so channels were introduced.

Notice that I didn't change the original `Walker` (concrete type `walker`); I
added a new type instead (`walkerC`) that implements the same `Walker`
interface. So you can actually have both implementations working and choose
among them when you build up the type in the `main()` function. This version is
now building the concurrent walker by running `NewWalkerC()` instead of the
original `NewWalker()`.

### Abort on errors

Those of you that followed what we did at the workshop may notice a difference
[here][walker:main-loop]. That's because I included another feature we hadn't
yet introduced at the workshop (and finally didn't at all due to time
constraints, sorry). At any point during tree traversal the application may find
an error; let's say a file cannot be opened or fully read. With the serial
version the tool was able to abort right away, because there was nothing else
running. However, things are a bit different for the concurrent version. When an
error is found at one routine, there might be others already running. Yet, we
don't want to wait for all of them to finish to simply tell the user that the
whole computation failed.

So instead of waiting for all routines, no matter their result, the walker
[aborts][walker:return] as soon as the first error is found. That's pretty much
the same thing we were doing with the serial version. Now think recursively.
That will make all pending invocations, all the way up to the root, abort
immediately. But even if we managed to exited the walker's `Walk()` function, we
still potentially left a lot of routines behind, running in background. Those
will not prevent the program from exiting, as we've already seen, but will be
needlessly wasting resources until they are done, if the tool had other things
to do and decided not to quit.

I've introduced another channel (`quitCh`) to fill that gap and ask pending
routines to abort. Recall that you cannot kill routines in Go; you have to be
polite and ask them to exit. (Whether they honor your request or not is of
course up to each routine.) The `select` statement you see [here][walker:select]
is a Go construction used to wait until one of the actions happens. Recall that
you read from a channel with `<-ch` and write with `ch <- msg`. Both operations
block until the action completes; `select` lets you wait on both until at least
one of them can proceed. You can list as many cases as required in the `select`
block, and mix sends and receives as appropriate. (There's also a `default`
clause that we're not using here, that prevents you from blocking if no case is
ready.)

Under normal conditions there would be nothing to read from `quitCh`, hence each
routine spawned from each walker instance will simply wait to send the result
and quit. Those results will be collected by [this][walker:collect] loop, that's
handling a specific directory in the hierarchy. But what if an error condition
was triggered? Given that we made the loop abort immediately (see the `return`
statement within), nobody else would be reading from the channel, and thus all
routines would block indefinitely trying to send their respective results.
There's where `quitCh` comes to the rescue. You may have noticed that I changed
the `Walk()` function not to be recursive itself, but to call a `doWalk()`
function that is, instead. This allows us to detect when we're coming out of the
outer-most level and do something special: [close][walker:close-quit] the
`quitCh`. When closed, readers waiting on the channel will be notified, and the
`select` we were discussing above would unblock.

Note that there's absolutely no need to close channels in Go. In fact, "close"
is probably a bad name for this operation. There's no memory leak, or things
like that, if you don't close a channel. Think of closing as a broadcast. If you
sent a single value, a single reader will get it, no matter how many of them are
waiting on the channel. However, by closing it, all readers will unblock
immediately. (Now you may want to look at the `select` again.) Beware not to
stretch the analogy with a broadcast too much; you can't "send two broadcasts"
by closing the channel twice. Do that and you'll get a panic. Also, you can no
longer write to a channel that's been closed. Do that and guess what... you'll
get a panic as well. You can still read from a closed channel, though, even if
you start reading way after the channel has been closed. A read from a closed
channel never blocks and always returns the zero for the type.

Side note: You may ask yourself what happens if the zero for the type is
actually a valid value in your domain. Answer: good question! You may use an
alternative syntax for receiving, that uses a second result to tell you that.
The statement `r, ok := <-ch` will read from channel `ch` and set the boolean
variable `ok` to `true` if and only if `r` was sent over the channel. When
`false` it means `ch` was closed and thus the zero value you'll get at `r` was
not really sent over the channel, but generated by Go instead. In our example
we're not even worried about the value, so there's no need for this.

Even if `quitCh` seems to be a good idea, checking it *after* the recursive call
has been made doesn't look like saving us a lot of work. We'd still have to wait
for the call to complete to check whether to abort or not, which potentially
implies a lengthy operation (think of checksums on huge files). But there's
fortunately another thing we haven't considered so far, that allows us to
introduce an extra check for `quitCh` before the checksum is even attempted.
(Aborting *during* checksum is out of scope for this demonstration. It adds
complexity, yet little value to the topics we're discussing.) Keep reading...

### File handles

By starting a recursive walk on a tree and concurrently processing every file we
find, we may easily hit the limits imposed by the OS regarding the maximum
number of open files. We didn't find that restriction during the workshop
because the directory I was using was pretty small. But that's something we
definitely need to take care of, so I took the chance to introduce another
technique here: buffered channels. (Wait... it was the other way around; this
example was though on purpose to have this problem so I could show you :))

Check [this][rsum:max-files] constant. That's the maximum number of files to
open at the same time. When built, the concurrent walker takes this number and
creates a buffered channel with that size, filling it with empty structures (see
[here][walker:tokens]). In this case it's not each item's value we're interested
in (hence the empty `struct{}`), but the fact that the value exists and uses one
slot in the channel, because we're regarding those as tokens. So before opening
a file, each routine will try to get one token and, as it's always the case for
channels, will block if none is available. Needless to say, we need to guarantee
that we'll return the token to the channel after the routine is done with the
file. 

At this point you may ask yourself why not keeping a counter, increasing it with
each open and decreasing it with each close. That's an option and you can do it,
but it's not "the Go way" to do it. Keep in mind that increasing and decreasing
here need to be atomic due to the multiple routines doing it concurrently. But
locking is precisely what we tend to avoid in Go, whose motto is "don't
communicate by sharing; share by communicating instead", and that's what it's
optimized to do.

Notice that we're using closures when [setting][walker:open] the `open()`
function for the `walkerC` structure. The `tokenCh` variable is local to the
constructor function (`NewWalkerC()`), yet it will be available for this inner
`open()` to use. Notice as well that this `open()` function will be used both
for opening files and directories, so the number we set is an absolute maximum;
there will be no point in time when we open more files than allowed.

Now what about the cancellation we were talking about before? Given that we need
to wait for a token before the file can be opened (and thus checksummed) we have
the chance to [use][walker:open-select] another `select` to keep an eye on our
old `quitCh` as well. So if one routine finds an error and the outer-most
`doWalk()` call returns, all waiting routines will end as well due to `quitCh`
being closed, even if they still haven't checksummed the file.

### Any updates to fdup?

Yes; a tiny but quite important detail. The fdup tool defines its own digesting
function for files (which, in turn, uses the standard digest we've been working
with). It turns out that this function needs a map to fill in with all digests
that are found in the tree. But maps are not safe for concurrent access in Go,
so we need to add synchronization. I've thus [added][fdup:mutex] a `Mutex`
variable that we need to [lock][fdup:mutex-lock] before reading/writing the map.
This is a standard use for locking, but if you don't want to use it, you can
spawn a separate routine and access the map only from there. The downside is
that you need to implement the set and get operations yourself, that will work
over the channel. That's usually not worth the trouble, so locking is widely
accepted for this.

### Thanks!

Finally, thanks for joining the workshop. So sorry time didn't permit exploring
further a lot of these things. Hope this text makes up for that :) Enjoy!


  [fdup:mutex-lock]:    https://github.com/gkristic/tech-meetup-2014/blob/316d4049c5c67a2383eb5a49f902dd33d97abebf/fdup/dups.go#L35-L36
  [fdup:mutex]:         https://github.com/gkristic/tech-meetup-2014/blob/316d4049c5c67a2383eb5a49f902dd33d97abebf/fdup/dups.go#L20-L23
  [rsum:main-loop]:     https://github.com/gkristic/tech-meetup-2014/blob/316d4049c5c67a2383eb5a49f902dd33d97abebf/rsum/main.go#L22-L50
  [rsum:max-files]:     https://github.com/gkristic/tech-meetup-2014/blob/316d4049c5c67a2383eb5a49f902dd33d97abebf/rsum/main.go#L12
  [walker:close-quit]:  https://github.com/gkristic/tech-meetup-2014/blob/316d4049c5c67a2383eb5a49f902dd33d97abebf/path/walker-c.go#L60-L62
  [walker:collect]:     https://github.com/gkristic/tech-meetup-2014/blob/316d4049c5c67a2383eb5a49f902dd33d97abebf/path/walker-c.go#L119-L125
  [walker:main-loop]:   https://github.com/gkristic/tech-meetup-2014/blob/316d4049c5c67a2383eb5a49f902dd33d97abebf/path/walker-c.go#L90-L115
  [walker:open-select]: https://github.com/gkristic/tech-meetup-2014/blob/316d4049c5c67a2383eb5a49f902dd33d97abebf/path/walker-c.go#L37-L33
  [walker:open]:        https://github.com/gkristic/tech-meetup-2014/blob/316d4049c5c67a2383eb5a49f902dd33d97abebf/path/walker-c.go#L30-L50
  [walker:return]:      https://github.com/gkristic/tech-meetup-2014/blob/316d4049c5c67a2383eb5a49f902dd33d97abebf/path/walker-c.go#L120-L123
  [walker:select]:      https://github.com/gkristic/tech-meetup-2014/blob/316d4049c5c67a2383eb5a49f902dd33d97abebf/path/walker-c.go#L110-L113
  [walker:tokens]:      https://github.com/gkristic/tech-meetup-2014/blob/316d4049c5c67a2383eb5a49f902dd33d97abebf/path/walker-c.go#L22-L25
