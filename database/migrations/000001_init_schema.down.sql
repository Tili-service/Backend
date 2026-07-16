-- On supprime l'historique en premier car il dépend de sales et profile
DROP TABLE IF EXISTS sale_history;

-- Ensuite on peut supprimer les ventes et les profils
DROP TABLE IF EXISTS sales;
DROP TABLE IF EXISTS profile;

-- On remonte l'arbre du catalogue
DROP TABLE IF EXISTS item;
DROP TABLE IF EXISTS categorie;
DROP TABLE IF EXISTS catalog;

-- On supprime le store et ses dépendances principales
DROP TABLE IF EXISTS store;
DROP TABLE IF EXISTS licence;
DROP TABLE IF EXISTS account;

-- Tables indépendantes (l'ordre importe peu pour celles-ci)
DROP TABLE IF EXISTS image;
DROP TABLE IF EXISTS payementmethod;