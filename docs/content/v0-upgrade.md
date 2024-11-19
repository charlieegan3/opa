---
title: Upgrading to v1.0
kind: documentation
weight: 122
---

All users should plan to upgrade to OPA v1.0 eventually. Some users, with more
control over the Rego loaded into the OPA instances they run will be able to do
so more quickly. Other users with less control or running third party Rego may
wish to upgrade to OPA v1.0 and use the v0 compatibility functionality to
upgrade gradually.

This documentation covers the different upgrade scenarios and the best course of
action for each. The documentation makes use of the following concepts:

- **Bundle Producer**: A system based on `opa build` that produces a bundle that
  is loaded by consumers.
- **Bundle Consumer**: An OPA instance running that loads and evaluates policy
  from a bundle in the system.
- **Authoring**: The process of writing Rego policies before bundles are
  produced and consumed. In managed systems, this might be a user's only contact
  point with OPA.

In some systems, where OPA is used without a bundle, there is no producer. This
simplifies the upgrade process.

## General Upgrade Approach: Upgrade Producers, then Consumers

Users will need to upgrade to OPA v1.0 in their own way, depending on their
release and change management processes, use cases and risk tolerance. The
general advice is to upgrade producers first, then consumers. Some users may
wish to migrate to OPA v1.0 all at once, with adequate testing and validation
this is possible. Not all steps are necessary for all users so a hybrid approach
is also an option depending on your context.

The rest of this documentation is designed to meet users where they find
themselves and direct them down the smoothest path to upgrade to OPA
v1.0.

## Detailed Producer & Consumer Version Scenarios

Tabulated in this section are the different versions of OPA users might be
working with in different parts of their systems. Select the scenario that best
matches your setup to find the recommended upgrade path.

If you are in doubt, [Scenario 1](#scenario-1) is the most common starting
point and we recommended you start there.

| OPA Versions  | Consumer v0.x                        | Consumer Versions Mixed   | Consumer v1.0                        |
| ------------- | ------------------------------------ | ------------------------- | ------------------------------------ |
| Producer v0.x | [Scenario 1](#scenario-1) (All v0.x) | [Scenario 4](#scenario-4) | [Scenario 7](#scenario-7)            |
| Producer mix  | [Scenario 2](#scenario-2)            | [Scenario 5](#scenario-5) | [Scenario 8](#scenario-8)            |
| Producer v1.0 | [Scenario 3](#scenario-3)            | [Scenario 6](#scenario-6) | [Scenario 9](#scenario-9) (All v1.0) |

<!-- source https://docs.google.com/drawings/d/137EObOVhMIVk9NEWOX0u_1eQOqe0MRkPQlmwKXkYWEU/edit -->

{{< figure src="opa-v0-upgrade.png" width="65" caption="Upgrade flows to OPA v1.0" >}}

## Upgrade Scenarios

Listed below are the different upgrade scenarios and the recommended migration
plans for each case.

### Scenario 1

All OPA runtimes - both bundle consumers and producers - are v0.x. This is the
most common starting point for users upgrading from a v0.x version of OPA.

#### How to Run

- Policies are authored to be v0.x compatible.

#### Next

Start upgrading producers to v1.0 ([Scenario 2](#scenario-2)) until all producers
are v1.0 ([Scenario 3](#scenario-3)).

### Scenario 2

Some bundle producers are v1.0, while some remain on v0.x. All bundle consumers
are v0.x. This might be the case if you have bundles from different tenants or
users using different OPA versions and you cannot control the versions they use.

#### Pre-requisites

Users cannot proceed with until they have either a single version of bundle
producers or have a means to control the use of `--v0-compatible` on newer
producers.

#### How to Run

- Policies are authored to be v0.x compatible.
- v0.x producers are run as-is.
- v0.x consumers are run as-is.

#### Next

Continue migrating producers to v1.0 until all producers have been upgraded
([Scenario 3](#scenario-3)).

### Scenario 3

OPA bundles are produced by OPA v1.0 instances, consumers are still on v0.x. This
scenario is common as users upgrade to OPA v1.0 by upgrading their producers
first.

#### Pre-requisites

Control of producers to set `--v0-compatible` or use `rego.v0` imports is
required.

#### How to Run

- Policies are authored to be v0.x compatible.
- v0.x consumers are run as-is.
- v1.0 producer is run with `--v0-compatible`, or modules have `rego.v0` import.
- Since policies will always be consumed by a v0.x OPA, all policies _must_ be v0.x compliant.

#### Next

Now that all producers are v1.0, and consumers are still not all v1.0, it's time
to get all the consumers to v1.0, ([Scenario 6](#scenario-6)).

### Scenario 4

Producers are v0, consumers are a mix of v0 and v1.0. This scenario might occur
when users have partially upgraded OPA instances to v1.0, but have not yet
upgraded their consumers. This is not a recommended step if it can be avoided as
it's recommended to upgrade producers first.

#### Pre-requisites

OPA v1.0 consumers must be able to run with `--v0-compatible` to accept v0
bundles. This upgrade path cannot continue until this is possible.

#### How to Run

- Policies are authored to be v0.x compatible.
- v0.x producers and v0.x consumers are run as is.

### Next

Downgrade consumers to v0.x ([Scenario 1](#scenario-1)) until all consumers are
v0. It's also possible to upgrade producers to v1.0 and continue the upgrade
from that point. Generally, it's recommended to upgrade producers first, however
depending on your existing OPAs v1.0 consumers deployments, you may prefer to
upgrade all your producers to v1.0 rather than to downgrade consumers.

### Scenario 5

Mixed versions of OPA are being used for both bundle production and consumption.

#### Pre-requisites

As users have a mix of bundle producers, they must have control over the runtime
options for the producers to set `--v0-compatible`. Users must also have control
over their v1 consumers to set the `--v0-compatible` flag. Both these conditions
must be met for the upgrade to proceed.

#### How to Run

- Policies are authored to be v0.x compatible.
- v1.0 consumers are run with `--v0-compatible`.

### Next

Please gradually upgrade producers to v1.0 ([Scenario 8](#scenario-8)) until all
producers are v1.0 ([Scenario 9](#scenario-9)).

### Scenario 6

All consumers can be run without flags, as the bundle will contain
attributes to inform v1.0 OPAs to accept v0.x modules.

#### How to Run

- Policies are authored to be v0.x compatible.
- v0.x consumers are run as-is. Bundles will contain v0.x policies
- v1.0 consumers are run as-is. Bundles will contain rego version attribute, so v0.x modules are accepted.

#### Pre-requisites

If users cannot set their OPA v1 producers to use `--v0-compatible` to be
compatible with their v0.x consumers, then this upgrade path is blocked.

#### Next

Running exclusively v1.0 producers and consumers, ([Scenario 9](#scenario-9)), is
the next and final step.

### Scenario 7

All consumers are v1.0, but producers are v0.x. This scenario might occur when
OPAs used for evaluation are upgraded before the policy bundling system.

#### Pre-requisites

If v1.0 consumers cannot be run with `--v0-compatible`, when loading v0 consumer
generated bundle, the bundles cannot include Rego version attribute. This means
the upgrade path is blocked until either the consumers can create bundles with a
Rego version or the `--v0-compatible` flag is available for producers.

#### How to Run

- Policies are authored to be v0.x compatible.
- v1.0 consumers are run with `--v0-compatible`

#### Next

Upgrade producers to v1.0 ([Scenario 8](#scenario-8)) until all producers are v1.0
([Scenario 9](#scenario-9)).

### Scenario 8

All consumers are v1.0, but producers are a mix of v0.x and v1.0.

#### How to Run

- Policies are authored to be v0.x compatible.
- v1.0 consumers are run with `--v0-compatible`
- v1.0 producers are run with `--v0-compatible`
- `rego.v0` imports cannot be used; rejected by v0.x bundle producers.

#### Pre-requisites

If using v0 bundles, it must be possible to use `--v0-compatible` on the bundle
producers in order for them to work in the v1 consumers.

Similarly, on the consumer side. If `--v0-compatible` cannot be set on the
consumers, the bundles from v1 producers will not be accepted. v0.x bundle
producers will produce v1.0-incompatible bundles as they will have no Rego
version attribute.

#### Next

Upgrade producers to v1.0 ([Scenario 9](#scenario-9)), completing the upgrade.

### Scenario 9

Once you have all consumers and producers running at v1.0 then you have
completed the upgrade to OPA v1.0. If you are using `--v0-compatible`
functionality, the next task is to upgrade the Rego loaded into OPAs to Rego v1.

Regardless of whether you are now upgrading your Rego, we encourage users to
use `opa check`, `opa check --strict` and to lint their Rego projects if you
have not already done so to identify issues.
