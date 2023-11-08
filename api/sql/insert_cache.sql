INSERT INTO cache_matrix
VALUES (?,?,?)
ON CONFLICT (from_sdk, to_sdk, count) DO 
UPDATE SET
(from_sdk,to_sdk,count) = (excluded.from_sdk,excluded.to_sdk,excluded.count);