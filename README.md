# Go load balancer

Go load balancer is a load balancer written in Golang. Inspired by https://codingchallenges.fyi/challenges/challenge-load-balancer

## Plan

I wanted to have such a roadmap to declare this project to be finished. Within the roadmap, I break down into multiple parts that each of part will consist of functionality to work on the project.

### Roadmap

**I. Basic functionality**

- [X] Able to load balance the traffic to the backend server using a round-robin algorithm.
- [ ] Able to perform regular health check on the backend server.
- [ ] Able to handle concurrency to pick a server.

**II. Advanced functionality**
- [ ] Able to configure the load balancer using multiple of algorithms option.
- [ ] Able to configure the load balancer server target using .yaml file.
