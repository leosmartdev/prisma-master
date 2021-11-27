let ldb = db.getSiblingDB('aaa');

ldb.createCollection("records");

ldb.createCollection("sessions");
ldb.sessions.createIndex({sessionId: 1}, {
    name: "sessionIdUnique",
    unique: true,
    sparse: true
});

ldb.createCollection("users");
ldb.users.createIndex({userId: 1}, {
    name: "userIdUnique",
    unique: true,
    sparse: true
});
