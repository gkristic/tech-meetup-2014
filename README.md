Concurrency in Go - Workshop
============================

This content will be used as the basis for a workshop on Go concurrency at
TechMeetup 2014. The applications will be updated throughout the workshop.

Two different tools are included in this package: rsum and fdup. The former
computes an SHA1 signature on files, pretty much like the standard sha1sum Linux
tool does. However, rsum can digest directories as well, hence providing a fast
check on whether arbitrarily deep subtrees contain exactly the same files with
the same contents. (This is based on hashing data, so the standard "warning"
applies: two identical streams will always hash to the same value, but the
converse is strictly not true. But it's unlikely enough.) So here's an example:

```
gkristic@hubble:sample> rsum readme
7a92f5fcf49eae62dcf1ada9ec113899f747c431  readme
gkristic@hubble:sample> sha1sum readme
7a92f5fcf49eae62dcf1ada9ec113899f747c431  readme
```

With rsum you can sum any number of files at once, no matter whether they are
regular files or directories (only readme is a regular file here):

```
gkristic@hubble:sample> rsum *
a5b6de83f4004bb942b565c37037dd9ba948da5e  deep
26c617bca42d1b305704812e0bf51107179f201a  hellos
7a92f5fcf49eae62dcf1ada9ec113899f747c431  readme
4c4f8674a44370580bc6d3c5bb38dfdafb9063e4  workshop
gkristic@hubble:sample> 
gkristic@hubble:sample> sha1sum *
sha1sum: deep: Is a directory
sha1sum: hellos: Is a directory
7a92f5fcf49eae62dcf1ada9ec113899f747c431  readme
sha1sum: workshop: Is a directory
```

If no arguments are provided, the tool will checksum the current directory:

```
gkristic@hubble:sample> rsum
5d6c0451f52671471644e1392e6933fe52a32fa8  .
```

The samples folder where these commands were run is included in this repository.
You can run the commands yourself to check these results.

The second tool in this bundle is fdup, that helps you detect where you're
spending your precious disk space in replicated contents. The tool will analyze
a whole tree (defaulting to the one starting at the current directory) and
report which files have exactly the same contents. It will also show you how
much (total) space you're using among all copies, like so:

```
gkristic@hubble:sample> fdup
Replicated contents (totals 72B) at:
  deep/nested/subtree/greeting
  hellos/hello1
  hellos/hello2
Replicated contents (totals 70B) at:
  deep/nested/subtree/readme
  readme
```

The output is sorted with the biggest (total) sizes at the top, making it handy
if you want to free up some space in your drive.
