# heyfyi
An example Golang Web Application

See me live at [hey.fyi](http://hey.fyi)!

## What is this?
When I first started developing websites in the Go programming language, I frequently wished that there was a complete example of a simple but complete web application that I could inspect.

heyfyi aims to fill the need that I had as a developer. This project is designed to be a complete example for the novice developer, featuring basics such as a secure accounts system, email transmission, in-depth templating, and a basic AJAX api for some web requests.

## How was it made?
The program was designed around the [gocraft](http://github.com/gocraft/web) mux and middleware package. 

It is built to the following architecture diagram:
![Architecture Diagram](https://github.com/kiwih/heyfyi/raw/master/run/files/HeyFyi_Architecture.png)

## Will it be developed further?
Absolutely. It is possible that there are mistakes, conceptual or otherwise, in the code. If you spot any, please submit an issue or a pull request!

In addition, as I do not have much experience with UI, there are many places where this could be improved.

## How do I build/run it?
You can `go get github.com/kiwih/heyfyi` this project. 

Once you have acquired it, navigate into the /run folder and `go build`.

At the moment, the server runs on localhost:3000, but you can change this with command line arguments.

As this program uses sqlite, you will need gcc to use cgo. If you are developing on windows, I recommend [mingw-64](http://sourceforge.net/projects/mingw-w64/) and not cygwin.
