# go task engine

A multi-coroutine task processing engine.

## What?

Recently, i am doing some work on handling multi-tasks in multi-coroutines.

There are many types of tasks, like A,B,C. The handle method is different if the type of task is different. When requests are coming, i need handle them in multi-coroutines.

## How?

This is similar to a factory although i never seen it myself.

Thinking it in the mind:

There is factory on somewhere. And there are some conveyors is rolling in this factory.
It should be like this:

There are some conveyors is rolling in a factory. 

There are some workers working on the conveyor belt. They handle the parts on the rolling conveyor belt. Different parts are placed on different conveyor belts

Before the factory shuts down, workers should process all the parts on the conveyor belt

## Future

It's not good enough yet, welcome issue or pr.

