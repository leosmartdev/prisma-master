#!/usr/bin/python3

from utilities.session import Session

session = Session().login('admin', 'admin')
session.create_user('standard', 'user', ['StandardUser'])
session.create_user('usermanager', 'one', ['StandardUser', 'UserManager'])
session.create_user('fleetmanager', 'one', ['StandardUser', 'FleetManager'])
session.create_user('incident', 'manager', ['StandardUser', 'IncidentManager'])
