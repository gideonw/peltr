# peltr

## Proto
### Spec
```
+--------+----------- N ----------+--+
|  cmd   |      message data      |CR|
+--------+------------------------+--+
|   8b   |        N bytes         |2b|
+--------+------------------------+--+
```
### hello
Intro command to set worker id and capacity

Server:
`hello000\n\r`

Worker:
`hello000id,10\n\r`

ID: id

Capacity: 10

### ping / pong
ping pong to ensure the worker is still working.

Server:
`ping`

Client:
`pong`

### assign
Assign a job to a worker

Server:
`assign00<gobdata>\n\r`

Data in this case will be a list of ids, endpoints, and rates.

Client:
`assign00<gobdata>\n\r`

Data in this case will be an acknowledgement of receipt.

### ready

---

### Design
```
Server                                Worker

   │                                    │                                      │
   │                                    │                                      │
   ├─────────────Hello─────────────────►│                                      │
   │                                    │                                      │
   │◄─────────────Identify──────────────┤                                      │
   │                                    │                                      │
   │                                    │                                      │
   │                                    │                                      │
   ├────────────Ping───────────────────►│                                      │
   │                                    │                                      │
   │◄───────────Pong────────────────────┤                                      │
   │                                    │                                      │
   │                                    │                                      │
   │                                    │                                      │
   ├────────────Assign────────────────► │                                      │
   │                                    ├──────────Job────────────────────────►│
   │◄─────────────Working───────────────┤                                      │
   │                                    │                                      │
   │                                    │                                      │
   │                                    │                                      │
   │                                    │                                      │
   │                                    │                                      │
   │                                    │                                      │
   │                                    │                                      │
   │                                    │                                      │
```