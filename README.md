# clgo [![Test](https://github.com/crewlinker/clgo/actions/workflows/test.yaml/badge.svg)](https://github.com/crewlinker/clgo/actions/workflows/test.yaml)

Opinionated but Re-usable Go libraries for the Crewlinker platform

## usage

- Setup the development environment `mage -v dev`
- Run the full test suite: `mage -v test`
- To release a new version: `mage -v release v0.1.1`

## backlog

- [ ] MUST include a mechanism to provide isolated schemas to tests, using a "versioned" migration strategy
      in a migraiton directory
- [ ] MUST include tracing, and re-add the test for contextual postgres logging (from the old 'back' repo)
- [ ] SHOULD add the Atlasgo github integration for checking migrations
- [x] SHOULD Allow configuration of the postgres application name to diagnose connections
- [x] SHOULD allow iam authentication to a database
