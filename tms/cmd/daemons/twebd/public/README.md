# HasPermissionForAction 

- Define action enum in the proto
- Use enum in ACTION variable (example, moc.Notice_ACK.String())
- security/Permission.go
    - Check for standard user first 
    - Then Admin
- security/service/Filter
    - add path to authorization filter

    