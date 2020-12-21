---
np: 000
title: Proposals
author: George Antoniadis <george@noodles.gr>
discussion: https://github.com/nimona/go-nimona/pull/537
status: Draft
category: Meta
created: 2020-12-20
---

# np000 - Proposals

## Simple Summary

A formal approach to proposing new features or updates for nimona.

## Problem Statement

Until now nimona's spec and protocol was very fluid in order to allow us to
quickly iterate without fear or braking things. This has resulted in a lack of
specifications and concrete documentation.

As we are moving closer to a more solid protocol we'll need to spend more time
designing updates in order to make sure everything is backwards compatible and
does not break apps.

In addition to this, we need a good way to help newcomers understand the various
parts of nimona and the rationale behind some of the design decisions.

## Proposal

Instead of trying to maintain specification docs, create distinct and standalone
proposals for each new feature or update.

Documentation will serve as a TL;DR and attempt to guide the reader through the
various proposals.

The proposal process itself is influenced by
[golang's](https://github.com/golang/proposal#design-documents) and
[ethereum's](https://eips.ethereum.org/EIPS/eip-1)
proposal processes.

### Index example

* Objects
  * np001 - Hinted objects
  * np003 - Consistent hashing
  * np007 - Streams
  * np008 - Feeds
* Network
  * np002 - Message based network
  * np004 - Enforce mTLS
  * np006 - Object relaying
* Discovery
  * np005 - Hyperspace

### Proposal template

Proposals should be written in markdown format and should be wrapped around the
80 column mark.

A [template](np-template.md) exists that provides the basic structure for the
proposal.

### Proposal Review

All discussions around proposals should be done on the discussion page mentioned
in the proposal in the form of comments. That will usually be the a Github
issue, or pull request.

A group of contributors will be holding “proposal review meetings” to make sure
that proposals are receiving attention from the right people, raising important
questions, pinging lapsed discussions, and generally trying to guide discussion
toward agreement about the outcome. The discussion itself is expected to happen
on the proposal's discussion page and not the meetings themselves.

The proposal review meetings also identify issues where consensus has been
reached and the process can be advanced to the next step (by marking the
proposal accepted or declined or by asking for a design doc).

### Consensus and Disagreement

The goal of the proposal process is to reach general consensus about the outcome
in a timely manner.

If general consensus cannot be reached, the proposal review group decides the
next step by reviewing and discussing the issue and reaching a consensus among
themselves. If even consensus among the proposal review group cannot be reached
(which should be exceedingly unusual), the arbiter (@geoah) reviews the
discussion and decides the next step.

## Open issues (if applicable)

Current specs need to be converted to NPs and taken through the review process.
