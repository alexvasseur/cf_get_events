# ABOUT

This Cloud Foundry CLI plugin reports on AI, tasks and service instance usage across foundation, org and space.

It is called *BCR* as in _Basic Consumption Report_
as it aims at helping organizations and platform team & leads doing accounting of usage and tracking value they get from their Cloud Foundry rollout.

( _cf_get_events_ is a historical upstream fork name and has nothing to do anymore with that project but in the early days the fork helped start from a golang source code querying the CF cloud controller - great work from ECSTeam)

This project is authored by a Pivotal employee and Cloud Foundry advocate under open source license terms.

# PCF versions

Plugin version 2.5 has introduced new option to include/exclude spaces using PCF metadata labels, which relies on CF API v3 and is available with PCF 2.5 and later.
For usage with PCF 2.4 or older please download and use plugin version 2.3

# USAGE

## Install

Simply install as a standalone cloud foudry CF CLI plugin.

Get a release from this project page https://github.com/avasseur-pivotal/cf_get_events/releases
Install with:
```bash
cf install-plugin -f ./bcr-plugin-osx
```
and verify installation with
```
cf plugins
plugin             version   command name                 command help
bcr                2.4.0     bcr                          Get Apps and Services consumption details
bcr                2.4.0     label-space                  Manage space level labels metadata
```

(You don't need golang to run it.)

## Usage

You can use each option separately.

The `--monthly` option reports on the PAS AppsManager usage report provided by Pivotal across the last 7 months.
The other options provides the *then current* usage and may be used for further exploration of *then current* usage.

Given the broad access, you must be login as CF admin or equivalent UAA role (cloud_controller.admin and uaa.admin for example)

```
Usage:  cf bcr [options]
    --monthly
	--ai
	--si
```

### Using metadata labels for AI and SI reporting

To filter based on metadata labels from space level use the optional `--label-space <label_selector>` option

```
Usage:  cf bcr [options]
	--ai --label-space <label_selector>
	--si --label-space <label_selector>
```

The plugin also provide the `label-space` command to read/write/search space level metadata label.
Please refer to [PCF metadata documentation](https://docs.pivotal.io/pivotalcf/2-6/adminguide/metadata.html) and [CF API label_selector](https://v3-apidocs.cloudfoundry.org/version/3.74.0/index.html#labels-and-selectors) format conventions for more details.

```
Usage:  cf label-space [options]
	(no argument)			shows labels for current space
	--write com.test/key=value	write label for current space
	--delete com.test/key		delete label for current space
	--search <label_selector>	search across all orgs & spaces
```

## Example: BCR

```
cf bcr --monthly --ai --si

https://api.system.domain
+------+-------+--------+--------+-----------------+-----------------+
| YEAR | MONTH | AI AVG | AI MAX | TASK CONCURRENT | TASK TOTAL RUNS |
+------+-------+--------+--------+-----------------+-----------------+
| 2018 |     4 |    103 |    117 |               0 |               0 |
| 2018 |     3 |     98 |    103 |               1 |               1 |
| 2018 |     2 |     94 |     96 |               0 |               0 |
| 2018 |     1 |     96 |    103 |               0 |               0 |
| 2017 |    12 |     99 |    116 |               2 |              45 |
| 2017 |    11 |     68 |    118 |               2 |              29 |
| 2017 |    10 |     44 |     54 |               0 |               0 |
+------+-------+--------+--------+-----------------+-----------------+


+-------------------------+-------------------------+----+---------------+------------------+---------------+----------------------------------------------------------------------+
|           ORG           |          SPACE          | SI | PIVOTAL MYSQL | PIVOTAL RABBITMQ | PIVOTAL REDIS |                            OTHER SERVICES                            |
+-------------------------+-------------------------+----+---------------+------------------+---------------+----------------------------------------------------------------------+
| ATM                     | prod                    |  7 |             1 |                  |             4 | p-service-registry:1,scheduler-for-pcf:1                             |
| ATM                     | test                    |  9 |             1 |                1 |             4 | scheduler-for-pcf:1,app-autoscaler:1,p-service-registry:1            |
| ATM                     | stage                   |  7 |             1 |                  |             4 | p-service-registry:1,scheduler-for-pcf:1                             |
| development             | xx                      |  0 |               |                  |               |                                                                      |
| development             | volumes                 |  1 |               |                  |               | nfs:1                                                                |
| microservices           | cnt-fortune-teller      |  5 |             2 |                  |               | p-service-registry:1,p-config-server:1,p-circuit-breaker-dashboard:1 |
| microservices           | sc-stream               |  1 |               |                1 |               |                                                                      |
| p-spring-cloud-services | instances               |  3 |               |                3 |               |                                                                      |
| Pivotal                 | mongodb                 |  1 |               |                  |               | p-mongodb:1                                                          |
...



+-------------------------+-------------------------+----------------------------------------------+-----+--------+--------------+
|           ORG           |          SPACE          |                     APP                      | AI  | MEMORY |    STATE     |
+-------------------------+-------------------------+----------------------------------------------+-----+--------+--------------+
| Pivotal                 | bluegreen               | attendees-concourse-v2                       |   1 |    700 | STARTED      |
| Pivotal                 | bluegreen               | attendees-concourse                          |   3 |    700 | STARTED      |
| Pivotal                 | springcloud             | fortune-service                              |   1 |    512 | STARTED      |
| Pivotal                 | springcloud             | fortune-ui                                   |   1 |    512 | STARTED      |
| Pivotal                 | mongodb                 | spring-music                                 |   1 |   1024 | STARTED      |
| Pivotal                 | docker                  | dockerapp                                    |   1 |     64 | STOPPED      |
| Pivotal                 | zipkin-metrics          | shopping-cart-svc                            |   1 |    768 | STOPPED      |
...
| system                  | system                  | p-invitations                                |   2 |    256 | STARTED      |
| system                  | system                  | app-usage-worker                             |   1 |   1024 | STARTED      |
| system                  | system                  | app-usage-scheduler                          |   1 |    128 | STARTED      |
| system                  | system                  | app-usage-server                             |   1 |    128 | STARTED      |
| system                  | system                  | apps-manager-js                              |   6 |    128 | STARTED      |
| system                  | system                  | apps-manager-js-venerable                    |   6 |    128 | STOPPED      |
| system                  | system                  | p-invitations-venerable                      |   2 |    256 | STOPPED      |
| system                  | system                  | app-usage-server-venerable                   |   1 |    128 | STOPPED      |
| system                  | system                  | app-usage-worker-venerable                   |   1 |   1024 | STOPPED      |
| system                  | system                  | app-usage-scheduler-venerable                |   1 |    128 | STOPPED      |
| p-spring-cloud-services | instances               | config-ad35392f-5e0f-410a-a6d9-63f63dc8f601  |   1 |   1024 | STARTED      |
| p-spring-cloud-services | instances               | eureka-87b64aef-fd93-4ae6-8dbe-ffe55eb38860  |   1 |   1024 | STARTED      |
+-------------------------+-------------------------+----------------------------------------------+-----+--------+--------------+
|            9            |           28            |                      77                      | 119 | 74206  | 52 (STARTED) |
+-------------------------+-------------------------+----------------------------------------------+-----+--------+--------------+
+-----------------------+-----+-----+--------+
|       CATEGORY        | APP | AI  | MEMORY |
+-----------------------+-----+-----+--------+
| Total                 |  77 | 119 |  74206 |
| Total (excl system)   |  47 |  56 |  48158 |
| STARTED               |  52 |  74 |  47966 |
| STARTED (excl system) |  34 |  38 |  32734 |
+-----------------------+-----+-----+--------+
+-------------------------+--------------+--------+----------------+
|           ORG           | MEMORY LIMIT | MEMORY | MEMORY USAGE % |
+-------------------------+--------------+--------+----------------+
| development             |        10240 |    750 |              7 |
| production              |        10240 |   3072 |             30 |
| microservices           |        10240 |   3584 |             35 |
| Pivotal                 |       102400 |   7920 |              7 |
| ProjetA                 |        10240 |      0 |              0 |
| ProjetB                 |        10240 |      0 |              0 |
| p-spring-cloud-services |       153600 |  15360 |             10 |
| qa                      |        10240 |   2048 |             20 |
| system                  |       102400 |  15232 |             14 |
+-------------------------+--------------+--------+----------------+
```

## Example: BCR with metadata labels for spaces

```
cf bcr --ai --si --label-space com.test/alex

https://api.system.domain
2.5.5-build.15 (Small Footprint PAS)
Filtering spaces with label selector: com.test/alex

+------+-------+----+---------------+------------------+---------------+----------------+
| ORG  | SPACE | SI | PIVOTAL MYSQL | PIVOTAL RABBITMQ | PIVOTAL REDIS | OTHER SERVICES |
+------+-------+----+---------------+------------------+---------------+----------------+
| Alex | dev   |  0 |               |                  |               |                |
+------+-------+----+---------------+------------------+---------------+----------------+
|  -   |   -   | 0  |       0       |        0         |       0       | 0 (PIVOTAL: 0) |
+------+-------+----+---------------+------------------+---------------+----------------+

+---------+------+-------------------+
| SERVICE | PLAN | SERVICE INSTANCES |
+---------+------+-------------------+
+---------+------+-------------------+

+------+-------+---------+----+--------+-------------+--------------+
| ORG  | SPACE |   APP   | AI | MEMORY |    STATE    | MEMORY USAGE |
+------+-------+---------+----+--------+-------------+--------------+
| Alex | dev   | session |  2 |   1024 | STOPPED     |              |
| Alex | dev   | booking |  1 |   1024 | STOPPED     |              |
+------+-------+---------+----+--------+-------------+--------------+
|  1   |   1   |    2    | 3  |  3072  | 0 (STARTED) |      0       |
+------+-------+---------+----+--------+-------------+--------------+

+------+--------------+--------------+---------+--------------+
| ORG  | MEMORY LIMIT | MEMORY USAGE | USAGE % | AI (STARTED) |
+------+--------------+--------------+---------+--------------+
| Alex |        32768 |            0 |       0 |            0 |
+------+--------------+--------------+---------+--------------+

+-----------------------+-----+----+--------+
|       CATEGORY        | APP | AI | MEMORY |
+-----------------------+-----+----+--------+
| Total                 |   2 |  3 |   3072 |
| Total (excl system)   |   2 |  3 |   3072 |
| STARTED               |   0 |  0 |      0 |
| STARTED (excl system) |   0 |  0 |      0 |
+-----------------------+-----+----+--------+
```

Other label_selector examples:

```
cf bcr --ai --si --label-space com.test/alex=dev

cf bcr --ai --si --label-space !com.test/alex=prod
...
```
Please refer to [PCF metadata documentation](https://docs.pivotal.io/pivotalcf/2-6/adminguide/metadata.html) and [CF API label_selector](https://v3-apidocs.cloudfoundry.org/version/3.74.0/index.html#labels-and-selectors) format conventions for more details.


## Example: read/write/search metadata labels for spaces

```
> cf target
api endpoint:   ...
api version:    2.131.0
user:           ...
org:            Alex
space:          dev

> cf label-space
com.test/alex=1933

> cf label-space --write com.test/alex=0014
com.test/alex=0014

> cf label-space --search com.test/alex=0014
+------+-------+--------------------------------------+--------------------------------------+
| ORG  | SPACE |               ORG GUID               |              SPACE GUID              |
+------+-------+--------------------------------------+--------------------------------------+
| Alex | dev   | a0054264-8010-4068-9305-9e6b459c972d | 488016a7-8c75-4c20-901a-3b80bc550693 |
+------+-------+--------------------------------------+--------------------------------------+

> cf label-space --search !com.test/alex
+----------------------------+------------------------------+--------------------------------------+--------------------------------------+
|            ORG             |            SPACE             |               ORG GUID               |              SPACE GUID              |
+----------------------------+------------------------------+--------------------------------------+--------------------------------------+
| system                     | system                       | 82f806ab-542f-4505-b0cc-e2163429e54b | 1055e417-d1da-437a-b89d-556296b8dcea |
...
```

## Uninstall

```
cf uninstall-plugin bcr
```
