-- A sample script
-- {{.Iter}} will be replaced by the current iteration count
BEGIN TRANSACTION;
INSERT INTO accounts (id, balance) VALUES({{.Iter}}, {{call .RandInt63}});
DELETE FROM accounts WHERE id = {{.Iter}}; 
COMMIT;