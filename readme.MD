# Go Event Bus

## Background
Experience using monolith application, we may find that we need to process a relatively big chunk of processes that need to be ran in the background. To address this problem, we need a process that can send a notification or message to trigger compute-heavy process. This problem statement was beautifully done by modern app e.g. Apache Kafka, NSQ, or using redis publish-subscribe. In the other hand, to use the mentioned applications we may need additional compute resource to address the problem. In the sense of efficiency and a green computing era, we may need more lightweight process with ability to send and consume message. This golang based event-bus is here leveraging the instance resource without deploying any other binaries and any data transfer overhead. 

This is an Event Bus abstract, before making use of this library, you may take a look to our features:

- Built with concurrency-processed-first in mind.
- Goroutine based event emitter
- Trigger and run the event listener using goroutine worker, so your listeners in a one event can ran independently and concurrently.
- Retryable listener, if the listener is returning an error, it will reattempt the process failure listener.
- Success and error callback, you can modify the callback based on your needs. You can run a DB call process, inserting a new row, emit another event, logs and etc.
