# dbwords

In preparation for a simple game that I want to create, I needed a database of words that are valid for play. I Googled up a tournament word list and quickly coded this little tool to do that. 

The only dependency this needs is the oh-so-excellent [BoltDB](https://github.com/boltdb/bolt) -- seriously, I love this stuff.

Utilizing only a single Goroutine, essentially `main`, this tool can take quite a long time to complete. To help speed things up this program takes advantage of Go's biggest strength: concurrency. Using 8,000 Goroutines it takes about a second to put 82,551 words into the database. Pretty awesome stuff, but YMMV!