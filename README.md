# ABOUT

This Cloud Foundry CLI plugin reports on AI, tasks and service instance usage across foundation, org and space.

It is called *BCR* as in _Basic Consumption Report_
as it aims at helping organizations and platform team & leads doing accounting of usage and tracking value they get from their Cloud Foundry rollout.

( _cf_get_events_ is a historical upstream fork name and has nothing to do anymore with that project but in the early days the fork helped start from a golang source code querying the CF cloud controller - great work from ECSTeam)

This project is authored by a Pivotal employee and Cloud Foundry advocate under open source license terms.

# USAGE

## Install

Install as a standalone cloud foudry CF CLI plugin
Get a release from this project page
Install with
```
cf install-plugin -f ./bcr-plugin-osx`
cf plugins
plugin             version   command name                 command help
bcr                2.1.0     bcr                          Get Apps and Services consumption details
```

(You don't need golang to run it.)

## Usage

You can use each option separately.

The `--monthly` option reports on the PAS AppsManager usage report provided by Pivotal across the last 7 months.
The other options provides the *then current* usage and may be used for further exploration of *then current* usage.

Given the broad access, you must be login as CF admin or equivalent UAA role (cloud_controller.admin and uaa.admin for example)

## Example

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



## Uninstall

```
cf uninstall-plugin bcr
```
