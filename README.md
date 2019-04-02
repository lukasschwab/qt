# qt

`qt` is a simple command-line torrent client. It can't show you ads.

<img width="643" alt="Screenshot of qt torrenting Sintel" src="https://user-images.githubusercontent.com/4955943/55372727-96f68f00-54b7-11e9-97fe-d871bffca409.png">

## Getting started

### Prerequisites

+ The Go Programming Language. See [https://golang.org/doc/install](https://golang.org/doc/install).
+ This project's source code.
  + **Recommended:** run `go get github.com/lukasschwab/qt`
  + Alternatively, clone this repository.

### Build `qt`

**Recommended:** in the project directory, run the following command.

```sh
$ make install
```

*Alternatively,* to build the binary in the project directory instead of in your `$GOPATH/bin`, you can use the following command. In this case, you may want to move the resulting `qt` executable into your shell path so you can run it as `qt`.

```sh
$ make build
```

## Usage

To download a torrent, run the `qt` executable:

```sh
$ # In the project directory, run:
$ ./qt 'magnet:...'
$ # Or, if qt is in your shell path:
$ qt 'magnet:...'
$ # For example, to torrent Sintel:
$ qt 'magnet:?xt=urn:btih:08ada5a7a6183aae1e09d831df6748d566095a10&dn=Sintel&tr=udp%3A%2F%2Fexplodie.org%3A6969&tr=udp%3A%2F%2Ftracker.coppersurfer.tk%3A6969&tr=udp%3A%2F%2Ftracker.empire-js.us%3A1337&tr=udp%3A%2F%2Ftracker.leechers-paradise.org%3A6969&tr=udp%3A%2F%2Ftracker.opentrackr.org%3A1337&tr=wss%3A%2F%2Ftracker.btorrent.xyz&tr=wss%3A%2F%2Ftracker.fastcast.nz&tr=wss%3A%2F%2Ftracker.openwebtorrent.com&ws=https%3A%2F%2Fwebtorrent.io%2Ftorrents%2F&xs=https%3A%2F%2Fwebtorrent.io%2Ftorrents%2Fsintel.torrent'
```

## Wishlist

+ Codebase cleanliness
  - [ ] Torrent update channel abstraction: no more timer checks in `main`.
+ Usability
  - [ ] Distribution as a binary.
  - [ ] Demo site.
