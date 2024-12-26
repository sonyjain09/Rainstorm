# RainStorm: A Hybrid Distributed File System and Stream Processing Framework

## Overview

RainStorm is a distributed stream-processing framework developed by combining features from multiple distributed system components, including:
- Distributed Group Membership.
- Hybrid Distributed File System (HyDFS).
- Real-time Stream Processing with RainStorm.

This system enables scalable, fault-tolerant, and efficient processing of real-time streams with exactly-once semantics. Built with the foundations of distributed group membership and file systems, RainStorm can handle failures gracefully while performing distributed computations.

## Distributed Group Membership
- A membership service that maintains a list of active machines in the system.
- Detects machine failures, joins, and leaves using a Ping-Ack failure detection algorithm with and without suspicion mechanisms.
- Ensures time-bounded completeness, updating membership lists across all nodes within 10 seconds.
- Logs events like node joins, leaves, and failures for debugging and querying.

## Hybrid Distributed File System 
- Consistent Hashing: Distributes file blocks across servers in a decentralized manner for scalability.
- File Operations:
    - create: Adds new files to HyDFS.
    - get: Retrieves files from HyDFS.
    - append: Appends data to existing files while ensuring client ordering.
    - merge: Ensures consistency across file replicas without losing client appends.
- Client-Side Caching:
    - Caches recently accessed files for faster reads.
    - Implements cache eviction to handle memory constraints.
- Fault Tolerance:
    - Handles up to two simultaneous node failures, with automatic data re-replication.
    - Ensures eventual consistency and ordering of operations across replicas.

## Stream Processing Framework 
