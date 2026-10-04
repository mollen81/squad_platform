-- Порядок обратный созданию: сначала таблицы со ссылками, потом те, на
-- которые ссылаются, иначе DROP упирается во внешние ключи.
DROP TABLE IF EXISTS team_members;
DROP TABLE IF EXISTS team;
DROP TABLE IF EXISTS users;
DROP TABLE IF EXISTS events;
