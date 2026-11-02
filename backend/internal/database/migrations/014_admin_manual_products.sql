-- M06.2 manual administrator-created catalog products.
-- Imported supplier IDs remain explicit; manual IDs use a reserved high range.
CREATE SEQUENCE admin_manual_product_id_seq AS BIGINT
    MINVALUE 1 START WITH 8000000000000000000 NO CYCLE;

DO $admin_product_sequence_setup$
DECLARE
    max_product_id BIGINT;
    next_product_id BIGINT;
BEGIN
    SELECT COALESCE(MAX(id), 0) INTO max_product_id FROM products;
    IF max_product_id >= 9000000000000000000 THEN
        RAISE EXCEPTION 'product ID space is too high to reserve a safe manual-product range';
    END IF;
    next_product_id := GREATEST(max_product_id + 1, 8000000000000000000);
    PERFORM setval('admin_manual_product_id_seq', next_product_id, false);
END;
$admin_product_sequence_setup$;

ALTER TABLE products
    ALTER COLUMN id SET DEFAULT nextval('admin_manual_product_id_seq'::regclass);
ALTER SEQUENCE admin_manual_product_id_seq OWNED BY products.id;

ALTER TABLE admin_product_revisions
    DROP CONSTRAINT admin_product_revisions_action_check,
    ADD CONSTRAINT admin_product_revisions_action_check
        CHECK (action IN ('created', 'updated', 'archived', 'restored'));
