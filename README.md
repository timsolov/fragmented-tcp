# fragmented-tcp
Implementation of fragmented tcp server and client.

# Requirements
- Go 1.17+

# Usage

`make build` - Build application (default goal)
`make test`  - Run tests

# Low level protocol
Each message from and to server consists of 2 parts.
- First 2 bytes is a length of packet;
- All other bytes is a packet.

Example in hex:
| Length (2 bytes) | Content (x bytes) | Decoded |
|------------------|-------------------|---------|
| 00 02            | 41 42             | AB      |
| 00 03            | 41 42 41          | ABA     |

# High level protocol
The protocol which should be used for communication between clients.
The protocol contatins always one or several octets of strings separated by space.
The first octet of protocol always is type of message on which application can build its behaviour.
Responses on each command will be `OK` or `ERROR`:

`OK <PARAMETER>` - means that command evaluated successfully.
    `<PARAMETER>` - dynamic parameter depending on kind of command. See each commands for details.

`ERROR <REASON>` - menas command has failed.
    `<REASON>` - what's happened.

## PING/PONG message (Keep Alive)
The Server sends broadcast message `PING` to All clients each minute.
The client should answer on `PING` message by  `PONG` message that means he's online.

## HI message.
When Client connects to the server he should send first hi message:

`HI <NAME>`

- `HI` is a command means authrization message;
- `<NAME>` is the name of the client or client's id. Not possible to use spaces in name.

There are reserved name for the Server broadcasting - `SYSTEM`. No one can take this name.

The response will be `OK <NAME>` or `ERROR <REASON>`.

## Incoming messages
Since client is authorized (see Welcome message) he can receive broadcast messages:

`MSG <FROM> <TEXT>`

- `MSG` is a command means incoming message;
- `<FROM>` is the name of the client who sent the message;
- `<TEXT>` is the text of message.

## List of connected clients
Each client can ask the server about clients' names connected to the server.

`CLIENTS`

The response will include `OK` message with param contains names of clients connected to the server separated by new line `\n` character.
Example:
```
Client1
Client2
Client3
```
Or an `ERROR` message with reason.

## Send a private message.
Each client can send a message to some another client.

`MSG <TO> <TEXT>`

- `MSG` is a command means sending a message;
- `<TO>` is the name of the client to whom current client send a message;
- `<TEXT>` is the text of message.

The response will be `OK <TO>` or `ERROR <REASON>`. 

# TODO

- The max length of packet should be limited to prevent memory leaks;
- Configurable timeouts on read packets;
- Configurable timeouts on write packets.