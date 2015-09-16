#!/bin/bash

## This is a sample file for ENV.sh
#
# You can use env file in order to run and/or test the code.
# The *ENV.sh files contain the necessary environment variables
# to run the project.
#
# You can have multiple env files, like: ENV.sh (main), TEST_ENV.sh (test
# environment) etc.
#
# I order to use the ENV file, for example TEST_ENV.sh, run:
#
# source TEST_ENV.sh


## pattern: type://user:password@host:port/dbname
export DATABASE_URL="postgresql://:manu@127.0.0.1:5432/skeleton"
export DB_TYPE="postgresql"
export HTTP_PORT=5000

export COOKIE_STORE_SALT = "9s7YD807h*&DHhihSD123434SASDD__xxxxxxxxxxxxxxxxxxxxxxx"

#export DEBUG=TRUE
#export API_KEY=<...>
