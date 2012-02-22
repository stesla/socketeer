# socketeer

For when you don't want to write another websocket server.

## Usage

The Socketeer requires no configuration. Once it's installed, all you
have to do is run it. You can configure its port and its default
line-ending with command-line options.

    $ socketeer -http :6060 -line "\r\n"

It supports two modes of sending data to the remote socket:
byte-oriented and line-oriented.

    ws://socketeer:6060/bytes?host=example.com&port=12345

The second is line-oriented. This supports an optional `lineEnding`
argument to override the default set at launch. Each frame sent over
the websocket will have the line ending appended to it when sent to
the remote host.

    ws://socketeer:6060/lines?host=example.com&port=54321&lineEnding=%0A

If host or port are missing, or there is a problem parsing the query
string, the websocket will be closed with code 4000. If the connection
to the remote host fails, the websocket will be closed with code 4001.
