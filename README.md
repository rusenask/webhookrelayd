# webhookrelayd 

Webhook relay is a SaaS where users can get a unique endpoint and use it to route external service webhooks into internal systems.

This repository is for an agent to perform "last mile" webhook delivery. Read more [here](https://webhookrelay.com). 

Some common use cases:

* For example if you want GitHub webhooks to invoke Jenkins build on code push.
* Relay Freshdesk events to internal Jira
* Relay Trello wehbooks to internal Jira/Jenkins


## Quick start

To start - register an account here https://webhookrelay.com/register. Then:
1. Create a bucket
2. Create an 'input' inside that bucket, it should look something like this "https://webhookrelay.com/v1/webhooks/1dbceb20-2626-48b3-ab89-d65dd81a9d07"
3. Use that _input_ as an endpoint inside the system that will __produce__ that webhook. Think of an _input_ as your personal "inbox". I would like to advise 
   using single _input_ per producer.
4. Create an _output_. Output basically defines destination (where you want incomming webhook(-s) delivered).
5. Go to tokens (https://webhookrelay.com/tokens) and generate a token for your agent
6. Start _webhookrelayd_:

    ./webhookrelayd -k 4c7cff17-8726-431b-a6bc-82d6def0bdc8 -s Xf9DgBjhHjqH