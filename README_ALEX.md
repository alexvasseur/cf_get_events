This was forked from https://github.com/ECSTeam/cf_get_events
using instructions from

http://www.personal.psu.edu/bam49/notebook/gopath-github-fork/

Fork the repository of interest on Github
Make a directory for the project in $GOPATH/src/github.com/ORIGINALOWNER/REPONAME
Clone YOUR fork into that directory git clone https://github.com/YOURNAME/REPONAME
Run go get github.com/ORIGINALOWNER/REPONAME to pull dependencies.
Add original repository as an upstream remote git remote add upstream https://github.com/ORIGINALOWNER/REPONAME if you have not already.


