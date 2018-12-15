-- Create table
\benchmark once \name init
CREATE TABLE dbbench_simple (id INT PRIMARY KEY, balance DECIMAL);

-- How long takes an insert and delete?
\benchmark loop \name single
INSERT INTO dbbench_simple (id, balance) VALUES({{.Iter}}, {{call .RandInt63}});
DELETE FROM dbbench_simple WHERE id = {{.Iter}}; 

-- How long takes it in a single transaction?
\benchmark loop \name batch
BEGIN TRANSACTION;
INSERT INTO dbbench_simple (id, balance) VALUES({{.Iter}}, {{call .RandInt63}});
DELETE FROM dbbench_simple WHERE id = {{.Iter}}; 
COMMIT;

-- Delete table
\benchmark once \name clean
DROP TABLE dbbench_simple;