-- Keep the PostgreSQL rating domain aligned with the catalog contract.
ALTER TABLE products
    ADD CONSTRAINT products_rating_upper_bound CHECK (rating <= 5);
