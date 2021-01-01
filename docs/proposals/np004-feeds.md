---
np: 004
title: Feeds
author: George Antoniadis (george@noodles.gr)
status: Draft
category: Objects
created: 2019-10-29
---

# Feeds

## Simple Summary

Feeds enable multiple peers that have been mandated by the same identity to
keep track and synchronize objects the create and receive.

## Problem statement

There is currently no way for peers to synchronize with other peers with similar
mandates the data they care about.

For example a group of applications of the same user don't have a way to make
sure they know of all the conversations the user is having access to between
all the different applications they are using.

## Proposal

Each peer has a number of "registered types", these are the data types the
application can consume or create.

For each of those types, a feed will be created and synchronized with peers
of the same identity.

![feed](./np004-feed.drawio.svg)
