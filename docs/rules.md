## Dave explains Rules vs DataCollectionRules

- FGM: what are the `rules: any[]` in Config Format ?
- Dave Roe: 
  - to answer the question directly: the reason we typed them as any in
    the spec is because they aren’t needed for the passive agent (which is all
    that is covered by the spec). It was supposed to  indicate “you will see this
    attribute in the config data but you can ignore it as it’s only needed for the
    active agent”.
  - ConfigOptions.rules is actually typed as Rule[] (see typings below)
  - To give you a proper explanation of rules vs data collection rules:
    - Rules were added to support the rules defined by the user in the dashboard 
      - ie. remediations/anomalies.
    - In the case of remediations, there is some config (one or more remediations)
      in the rule which tells the agent how to behave (eg. set a timeout with
      duration X, retry for X attempts with Y delay). We also send/report back a
      rule id so we can match it up to the user-created rule on the server end
      (sent in activeRules, not in the agent spec).
    - We then added Data Collection Rules as a different type of rule to support
      some other features. These are not tied to a user-defined rule (no rule id)
      and initially had no config attached to them. Their purpose was to tag the log with some opaque server-supplied
      data (ie. report back the params attribute when the rule matched).
    - We later added config to data collection rules so they now supply some
      config and send params back.
  - Both rule types have now evolved to do similar things and, at some point
    in the future, I would like to change it like this:
    - Split config out of DataCollectionRule into a new ConfigRule type. So we 
      then have one kind of rule for receiving config and a separate rule type
      for tagging/collecting/sending data back.
    - Add options to data collection rule’s DynamicConfig to support the config
      currently defined by remediations from Rules
    - For each Rule the server is currently sending, send one ConfigRule with the remediations and one DataCollectionRule
      with the rule id in it’s params.
    - Drop Rules
    
Hopefully that made some sense to you! I thought it might be useful to describe 
the history/future to understand the present types.

```typescript
export enum RemediationType {
  BlockRequest = 'BlockRequest',
  Timeout = 'TimeoutRemediation',
  Retry = 'RetryRemediation'
}
export type BlockRequestRemediation = {
  typeName: RemediationType.BlockRequest
}
export type TimeoutRemediation = {
  typeName: RemediationType.Timeout
  duration: number
}
export enum RetryType {
  Exponential = 'EXPONENTIAL',
  Uniform = 'UNIFORM'
}
export type RetryRemediation = {
  typeName: RemediationType.Retry
  retryType: RetryType
  attempts: number
  delay: number
}
export type Remediation = BlockRequestRemediation | TimeoutRemediation | RetryRemediation
export interface Rule {
  id: string
  filterHash?: string
  remediations: Remediation[]
  expiresIn?: number
}

```

Regarding rules specifically, we don’t want to make the rules part of the public interface of the agent.
This is because:

- The agent rules aren’t in direct correspondence with the dashboard rules 
- we don’t want to have to document a completely separate rule system and have
  added complexity/confusion for the user.
- We’ll have to commit to keeping the same filter hash algorithm and implementing
  that in the agent, or we can’t merge safely.
- The current format of rules isn’t particularly logical/clear. We don’t want to 
  have to support that (both in terms of compatibility and having to respond to
  support tickets).

We envisage the user managing the rules centrally - using the platform means you don’t need to change/redeploy your app
to make changes. And the user-defined/dashboard rules can be dynamic in nature (we analyse a time window of  historic
data and add/remove agent rules according to the results)

I think merging (or rather replacing/overriding) config on the agent side, and
not exposing the rules/filters to the user directly, is the best choice for now.
