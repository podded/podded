# PODDED

## SDE

### TODO: CHANGE THIS OVER TO HIGGS

The following sde tables are expected to present

- invTypes.sql.bz2

```shell
wget https://www.fuzzwork.co.uk/dump/latest/invTypes.sql.bz2
wget https://www.fuzzwork.co.uk/dump/latest/invGroups.sql.bz2
wget https://www.fuzzwork.co.uk/dump/latest/invCategories.sql.bz2
bzip2 -dc invTypes.sql.bz2 | docker-compose exec -T mariadb sh -c 'mysql -h localhost -u podded -ppodded podded'
bzip2 -dc invGroups.sql.bz2 | docker-compose exec -T mariadb sh -c 'mysql -h localhost -u podded -ppodded podded'
bzip2 -dc invCategories.sql.bz2 | docker-compose exec -T mariadb sh -c 'mysql -h localhost -u podded -ppodded podded'
```

## Cleaner

This module will clean up the raw esi data and try and account for any changes over time that occur.

The main one at the moment is the systems changing to triglavian space in the povchen region that occurred 2020-10-13.

The cleaner requires that in the podded database in mariadb, the following tables exist
- map_solar_systems_2020_10_08
- map_solar_systems_jumps_2020_10_08
- map_solar_systems_2020_10_13
- map_solar_systems_jumps_2020_10_13

These tables are seeded from Fuzzy Steve's SDE conversions mapSolarSystems.sql.bz2

The former comes from sde-20201008, the latter from sde-20201013

The following script will download the two required files and add them to the database

```shell
#!/usr/bin/env bash

# These are the systems
wget -O map20201013.sql.gz https://www.fuzzwork.co.uk/dump/sde-20201013-TRANQUILITY/mapSolarSystems.sql.bz2
wget -O map20201008.sql.gz https://www.fuzzwork.co.uk/dump/sde-20201008-TRANQUILITY/mapSolarSystems.sql.bz2

# These are the jumps
wget -O jumps20201013.sql.gz https://www.fuzzwork.co.uk/dump/sde-20201013-TRANQUILITY/mapSolarSystemJumps.sql.bz2
wget -O jumps20201008.sql.gz https://www.fuzzwork.co.uk/dump/sde-20201008-TRANQUILITY/mapSolarSystemJumps.sql.bz2


bzip2 -dc map20201013.sql.gz | docker-compose exec -T mariadb sh -c 'mysql -h localhost -u podded -ppodded podded'
bzip2 -dc jumps20201013.sql.gz | docker-compose exec -T mariadb sh -c 'mysql -h localhost -u podded -ppodded podded'
# NOW RUN THE FOLLOWING SQL
# RENAME TABLE mapSolarSystems TO map_solar_systems_2020_10_13, mapSolarSystemJumps TO map_solar_systems_jumps_2020_10_13;

bzip2 -dc map20201008.sql.gz | docker-compose exec -T mariadb sh -c 'mysql -h localhost -u podded -ppodded podded'
bzip2 -dc jumps20201008.sql.gz | docker-compose exec -T mariadb sh -c 'mysql -h localhost -u podded -ppodded podded'
# NOW RUN THE FOLLOWING SQL
# RENAME TABLE mapSolarSystems TO map_solar_systems_2020_10_08, mapSolarSystemJumps TO map_solar_systems_jumps_2020_10_08;




```