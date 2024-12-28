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
- Design Summary: Each machine maintains a full and accurate membership list, pinging a randomly selected node every 0.5 seconds. If no acknowledgment is received within 1.5 seconds, the node is flagged and further action is
taken based on the suspicion mechanism. We created an introducer machine to handle new machines joining the system. It listens for join requests, adds new nodes, and shares the updated membership list with all new machines that join. If the introducer is
down, no new nodes can join, but existing processes continue as normal. In the basic Ping-Ack process, if no acknowledgment is received, the node is immediately removed, and a failure message is broadcast to the system for quick detection. However, this can
lead to inaccuracies, which are fixed by Ping-Ack+S. In the Ping-Ack+S implementation, a missed acknowledgment places the node in a "suspected" state for 8 seconds, instead of marking it as failed. All machines are notified and update their lists. The
node increments its incarnation number during suspicion. If it responds within 8 seconds to any machine, its status is set back to "alive" with an updated incarnation number. Otherwise, it is marked as "failed" on all machines. Failure information is
broadcast immediately to all machines. The 1.5-second timeout quickly flags unresponsive nodes as suspicious, while the Ping-Ack+S mechanism waits 8 seconds before marking a node as failed, allowing time for recovery from network issues.

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

## Setup
1. Clone the repository onto any machine that will be added to the distributed system.
2. Set appropriate variables
- In the global/global-vars.go file there are 3 variables that need to be set:
  - Tcp_ports
  - Udp_ports
  - Rainstorm_ports
    
Each of these is a list of addresses (hostname and port) that are part of the distributed system. The hostnames should be the same across all of the lists. For best functionality, make all of the ports different across the lists.
- In the .env file:
  - TCP_PORT = machines own tcp port value
  - UDP_PORT = machines own udp port value
  - RAINSTORM_PORT=machines own rainstorm port value
  - MACHINE_NUMBER = assign each machine a number, based on its index in the created lists above  
  - MEMBERSHIP_FILENAME = debugging log for membership functionality
  - HYDFS_FILENAME = debugging log for hydfs functionality
  - RAINSTORM_FILENAME = debugging log for rainstorm functionality
  - MACHINE_UDP_ADDRESS = machines own udp address value
  - MACHINE_TCP_ADDRESS = machines own tcp address value
  - MACHINE_RAINSTORM_ADDRESS = machines own rainstorm address value
  - LEADER_ADDRESS = the address of the leader (this should be common across all machines and the port should be a rainstorm port)
  - INTRODUCER_ADDRESS = the address of the introducer (this should be common across all machines and the port should be a udp port)
3. cd into the repository and run the command "go run main.go" on all machines
4. To join the distributed system type "join"
5. Here are a list of the other commands you can use 
    - list_mem_ids
      
        List all machines part of the distributed system. This includes their unique node id, hash value, incarnation number and alive/suspicion status.
    - leave
      
        Leave the distributed system, with the ability to join back again later with the same node id.
    - enable_sus
      
        Enable the suspicion mechanism.
    - disable_sus
      
        Disable the suspicion mechanism.
    - status_sus
      
        View the current status of the suspicion mechanism.
    - list_sus
      
        List all nodes in the membership system that are currently suspected.
    - get <hydfs_filename> <local_filename>
    
        get a file named **hydfs_filename** from hydfs and store its contents in **local_filename**
    - create <local_filename> <hydfs_filename>
    
        create a file named **hydfs_filename** in hydfs using the contents from **local_filename**
    - append <local_filename> <hydfs_filename>
    
        append the contents of **local_filename** into the hydfs file **hydfs_filename**
    - merge <hydfs_filename>
    
        merge the contents of **hydfs_filename** across all replicas in all machines
    - ls
      
        list the hydfs files stored on the machine 
    - rainstorm <op1_exe> <pattern> <op2_exe> <stateful> <hydfs_src_file> <hydfs_dest_filename> <num_tasks>
    
        perform a streaming task with 2 executables. You have the ability to put in a pattern and specify whether its stateful or not. State which **hydfs_src_file** will provide the input and the name of the file
        for the output of the streams (**hydfs_dest_filename**). Specify the number of tasks (**num_tasks**) desired for each stage in the job.
## Contributors

Sonam Jain and Sai Aitha
