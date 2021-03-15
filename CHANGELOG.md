# Wharf GitHub plugin changelog

This project tries to follow [SemVer 2.0.0](https://semver.org/).

<!--
	When composing new changes to this list, try to follow convention.

	The WIP release shall be updated just before adding the Git tag.
	From (WIP) to (YYYY-MM-DD), ex: (2021-02-09) for 9th of Febuary, 2021

	A good source on conventions can be found here:
	https://changelog.md/
-->

## v0.1.1 (2021-03-16)

- Added CHANGELOG.md. (!2)

## v0.1.0 (2020-10-19)

- Added Go module. (!1)

- Added README.md (!1)

- Added first implementation towards RabbitMQ with the following key
  functions: (!1)

  - `NewConnection`
  - `MQConnection.Connect`
  - `MQConnection.CloseConnection`
  - `MQConnection.PublishMessage`

