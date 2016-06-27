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

Once you have acquired it, `go get -u` then `go build`.

As this program uses sqlite, you will need gcc to use cgo. If you are developing on windows, I recommend [mingw-64](http://sourceforge.net/projects/mingw-w64/) and not cygwin.

Finally, you can run it by running `./heyfyi`

On the first run, it will create the database file and tables. 

This will contain a default fact and a default admin user.

You can sign in to the default admin user with username/password both `test@test`.

## Setting environment variables

`$COOKIE_STORE_SALT` - used in cookie encryption. It defaults to `SUPER_SECRET_SALT` for testing purposes only.

`$HTTP_PORT` - set the HTTP port to listen on. Defaults to `3000`.

`$LOG_FILE_NAME` - set the logfile name for logging. Defaults to `heyfyi.txt`.

## Screenshots

![Screenshot 1](https://github.com/kiwih/heyfyi/raw/master/run/files/screen1.png)

![Screenshot 2](https://github.com/kiwih/heyfyi/raw/master/run/files/screen2.png)