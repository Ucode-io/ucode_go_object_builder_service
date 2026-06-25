DO $$
BEGIN
    UPDATE chats SET type = 'ucode';
EXCEPTION
    WHEN others THEN NULL;
END
$$;