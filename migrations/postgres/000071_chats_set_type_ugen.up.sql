DO $$
BEGIN
    UPDATE chats SET type = 'ugen';
EXCEPTION
    WHEN others THEN NULL;
END
$$;
