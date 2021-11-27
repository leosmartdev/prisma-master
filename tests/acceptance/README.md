# Integration Test Suite
  
## Overview
 
The Integration Test Suite will be testing all of the services that belong to the C2 API.

### Config file

The integration test suite has it's own configuration file that lives in the root of the `acceptance` tests directory. There are a few differences from what its in the `tmsd.conf` file we use in the development environment. The configuration file for the test suite contains the following:

```linux
users=vagrant
{tgwad --num 20 --name test}
{tsimulator --fstations /vagrant/tests/acceptance/stations_data.json
            --fvessels /vagrant/tests/acceptance/vessels_data.json
            --webaddr :8089
}
{tmccd --protocol tcp}
{tfleetd}
{tdatabased}
{twebd}
{tanalyzed}
{tauthd}
```

### What's the difference?

the difference between the `tmsd.conf` in the development environment is the following:

```linux
{tplayd /var/lib/trident/sensorData/css-btm-24hour 3452}
{tnoid --address :3452 --radar_latitude 1.187700 --radar_longitude 104.021866}
{tnoid --address :9011 --name tnoid3}
{tnoid --address :9012 --name tnoid4}
{tnoid --address :9001 --name tnoid5}
```

the above parameters are not in our test `tmsd.conf` file, the integration test suite does not test `the playback data` or `tforward`.

### Makefile 

*Note:* You can run the entire integration test suite from the BASEDIR by running the following task:

```bash
make integration
```

When you run `make integration` from the prisma directory the task will do the following:

```bash
1. The first thing does is turn off all HTTPS warnings when running the integration tests.
2. The second line goes to the following path and runs the `test-tmsd` bash script.
3. The third line is a `sleep` statement, which basically gives time for all of the daemons and database to come up properly.
4. And last but not least the fourth goes to the root of the `acceptance` tests directory and runs the suite.
```

### Database Utility

The PRISMA C2 integration test suite has its own separate database. The reason we have a separate database is because we wanted our test suite to run on an empty data-set (meaning no pre-populated data set). 

An example of that would be our batam data-set in the development environment. Each scenario starts off with a blank slate, and ends with an empty database. The new database is created when the `test-tmsd` script is called. The `test-tmsd` script is doing the following:

```linux
1. The first thing the task does is stop all of the daemons.
2. The second line will delete the trident db.
3. Then it runs the `start-db` bash script:
* The start-db script:
  * creates a new database called "db-it" (which stands for the database for integration tests)
  * which will be listening in on "port: 8201"
4. And last but not least the fourth goes to the root of the "acceptance" tests directory and runs the suite.
```

## PyMongo

We are importing the PyMongo library to do the following in our integration test suite:

1. Dropping collections in mongoDB when a `test_method` is done asserting the JSON being returned. The code is documented in the following class:

```python
class clean_collections_after:
```

## Usage 

This "class" will be used as a decorator, it will take a list of collection names and clear those collections using mongo.delete_many() when the wrapped function returns.

## Example

```python
@mongo.@clean_collections_after('collection1', 'collection2', ...)
```

When you import the mongo utility in your test class you will be able to add this decorator above the `test_method` to clean the collection after each individual test run. you can pass the collection(s) name as an argument in the decorator.  

```python
*Clear the fleets collection
@mongo.@clean_collections_after('fleets')

* Clear the fleets and vessels collection
@mongo.@clean_collections_after('fleets', 'vessels')
```

2. Importing data in mongoDB collections, before the block of code in the `test_method` is run. The code is documented in the following class:

```python
class insert_collection_data:
```

## Usage 

The class first drops any data that is in the collection(s), then imports data which is added to the collection(s) then runs the `test_method` which then asserts the JSON response then drops the data that is in the collection(s).

## Example 

```python
@mongo.insert_collection_data('collection1', 'collection2', ...)
```

When you import the mongo utility in your test class you will be able to add this decorator above the `test_method` to insert data in the collection before each individual test run. you can pass the collection name as an argument in the decorator.  

```python
*Insert data into the fleets collection
@mongo.insert_collection_data('fleets')
```

The Database Utility lives here:
```bash
utilities/mongo.py
```

# Running Tests

### Running the entire test suite

When all daemons have been invoked you can run the entire test suite from the `acceptance` dir by running the following command:

```bash
./run-tests
```

### Running a test class with test method(s)

A test class is created  which subclasses `unittest.TestCase` which then allows you to create individual test scenarios which are referred to as `test_methods` whose names must start with `test` (to run an entire test class you run the following command):

```bash
./run-tests fleets.test_create_fleet
```

The above command is: 

```python
Running the test_create_fleet.py file which is inside the fleets directory. The test_create_fleet.py file contains the class CreateFleet(unittest.TestCase):
```

The `CreateFleet(unittest.TestCase):` class has nine individual `test_methods`

### Running one single test method

```bash
./run-tests fleets.test_create_fleet.CreateFleet.test_with_valid_input_returns_success

The above command is doing the following:

Running
```

### Read the docs

* [unittest](https://docs.python.org/3.6/library/unittest.html)
* [requests for humans](http://docs.python-requests.org/en/master/)
* [pymongo](http://api.mongodb.com/python/current/tutorial.html)
