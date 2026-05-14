// scripts/mongo-init.js
// Exécuté une seule fois à la création du conteneur mongo.
// Crée un utilisateur dédié à l'application avec droits limités à la DB lambare.

const dbName = process.env.MONGO_INITDB_DATABASE || "lambare";
const appUser = process.env.MONGO_APP_USER || "lambare_app";
const appPassword = process.env.MONGO_APP_PASSWORD || "changeme_app";

db = db.getSiblingDB(dbName);

db.createUser({
  user: appUser,
  pwd: appPassword,
  roles: [{ role: "readWrite", db: dbName }],
});

// Collections initiales avec validation de schéma basique
db.createCollection("users", {
  validator: {
    $jsonSchema: {
      bsonType: "object",
      required: ["email", "password", "role"],
      properties: {
        email: { bsonType: "string" },
        password: { bsonType: "string" },
        role: { bsonType: "string", enum: ["admin", "user"] },
      },
    },
  },
});

db.createCollection("items");

print(`✓ Database '${dbName}' initialized with user '${appUser}'`);
