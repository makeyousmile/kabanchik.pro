const dbName = process.env.MONGO_INITDB_DATABASE || "kabanchik";
const appUser = process.env.MONGO_APP_USERNAME || "kabanchik_app";
const appPass = process.env.MONGO_APP_PASSWORD || "kabanchik_app_pass";

const db = db.getSiblingDB(dbName);

db.createUser({
  user: appUser,
  pwd: appPass,
  roles: [
    { role: "readWrite", db: dbName }
  ]
});
