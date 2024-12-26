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
- Design Summary: Each machine maintains a full and accurate membership list, pinging a randomly selected node and a subset of 4 predefined nodes every 5 seconds. If no acknowledgment is received within 1.5 seconds, the node is flagged and further action is
taken based on the suspicion mechanism. We created an introducer machine to handle new machines joining the system. It listens for join requests, adds new nodes, and shares the updated membership list with all new machines that join. If the introducer is
down, no new nodes can join, but existing processes continue as normal. In the basic Ping-Ack process, if no acknowledgment is received, the node is immediately removed, and a failure message is broadcast to the system for quick detection. However, this can
lead to inaccuracies, which are fixed by Ping-Ack+S. In the Ping-Ack+S implementation, a missed acknowledgment places the node in a "suspected" state for 8 seconds, instead of marking it as failed. All machines are notified and update their lists. The
node increments its incarnation number during suspicion. If it responds within 8 seconds to any machine, its status is set back to "alive" with an updated incarnation number. Otherwise, it is marked as "failed" on all machines. Each machine is pinged by at least
4 others every 5 seconds, ensuring fast failure detection, even with up to 3 simultaneous failures. Failure information is broadcast immediately to all machines. The 1.5-second timeout quickly flags unresponsive nodes as suspicious, while the
Ping-Ack+S mechanism waits 8 seconds before marking a node as failed, allowing time for recovery from network issues.

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
- Design Summary: Each machine is assigned a random hash ranging from 0 to 2048, creating a virtual ring. The hashing algorithm is deterministic ensuring any machine can get the hash of any string value. When a file enters the system, it is given a unique hash
and stored on the nearest machine with a higher hash value. This hash tells machines where a file is located. Given that data stored in HyDFS is tolerant of up to two simultaneous failures, we chose to have 3 replicas for each file. We determined that 3
replicas is the minimum number that guarantees data preservation in a worst-case scenario. Each file is labeled with its primary successor, indicating its origin. We utilize a pull-based approach when handling re-replication. Any time a node joins or leaves
the system each machine assesses if a predecessor or successor joined/left. Based on this, the virtual ring is used to send pull messages to other machines it needs files from. Excess replicas are also removed from any machine. To decrease the latency of
get requests, we implemented a cache. Any time a client performs a get request of a file, it is stored in a local cache, and the name of the file is stored in a map data structure. Any subsequent gets will first try to retrieve the file from the cache
before making any requests to other machines. If an append is called on a file in the cache, it is removed to ensure any get requests are returning consistent data. When a client performs an append/get to a file in the system, it will always access the
same replica based on the virtual ring. During an append, a ‘chunk’ file is added to the replica. This ensures that running “get” on any client will reflect its latest appends. Timestamps are also created and attached to each append to preserve proper ordering.
The merge function collects appends from each replica, orders them by timestamp, combines them into a single file, and distributes the merged file back to each replica, ensuring consistency across all replicas.

## Stream Processing Framework 
- Real-Time Analytics: Processes continuous streams of data without waiting for batch completion.
- Stream Operators:
  - Transform: Applies user-defined functions to input records.
  - FilteredTransform: Filters and transforms records based on conditions.
  - AggregateByKey: Maintains state and aggregates records by key.
- Fault Tolerance:
  - Implements exactly-once delivery semantics.
  - Maintains logs of processed records and state for recovery across failures.
- Data Sources:
    - Integrates with HyDFS for input data and logs.
    - Supports continuous data streaming for real-time processing.
- Scalability:
    - Dynamically schedules tasks across worker nodes.
    - Handles worker failures and reschedules tasks seamlessly.
- Design Summary: Rainstorm uses a leader server that is responsible for managing task scheduling across worker servers whenever a Rainstorm job is executed. The leader divides each stage into num_tasks and iterates through a queue of workers, assigning
tasks to them. Once the task schedule is finalized, the leader distributes it to all workers. Additionally, the leader sends the initial file partitions to the designated "Source" servers, which are responsible for dividing the file into individual lines.
Upon receiving the schedule, workers utilize it to process incoming data streams. For each processed stream, the worker determines the next destination based on the schedule and forwards the stream accordingly. Workers assigned to the final stage are
responsible for sending the processed output back to the leader and writing the final results to the destination file. To ensure robustness, each task is associated with an append-only log file. When a worker receives a tuple, it processes it and checks its
append-only file to determine if the tuple has been processed before. If the tuple is new, the worker processes it as normal. The sender logs the ack in its own append-only file. If a tuple has already been processed, the worker simply skips it. In the event of
a machine failure, the leader assigns replacement workers to take over the failed tasks. The new workers consult the append-only logs to identify which tuples require processing or retransmission. This mechanism also ensures that no tuple is processed more
than once. To reduce the overhead of frequent communication over TCP, Rainstorm batches streams, acknowledgments, and writes to HyDFS. Every 100 milliseconds, a batch of data is sent to its respective destination, significantly improving efficiency. For stateful
operations, workers maintain a local file that records the state of processed tuples. This state is persisted to HyDFS every 100 milliseconds. In the event of a failure, the persisted state log allows the server to resume from where the failed process left off.
