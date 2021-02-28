# whook

Whook creates a http server that listens on a topic and prints whatever is posted to stdout.

In order to subscribe to a topic pass argument to whook as follows:

    $ whook topic1 topic2

You can use topic url as webhook, example topic url: http://localhost:8000/topic1

After a message is posted to a topic url it will be printed to stdout
