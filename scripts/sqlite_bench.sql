BEGIN TRANSACTION;
INSERT INTO accounts (id, balance) VALUES( {{.Iter}}, {{.Iter}} );
DELETE FROM accounts WHERE id = {{.Iter}}; 
COMMIT;