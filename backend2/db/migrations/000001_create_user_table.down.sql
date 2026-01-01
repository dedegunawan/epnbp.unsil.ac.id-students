-- Hapus tabel pivot terlebih dahulu (untuk hindari error foreign key)

DROP TABLE IF EXISTS user_roles;
DROP TABLE IF EXISTS role_permissions;

-- Lalu tabel utama (harus urut agar tidak konflik dependensi)

DROP TABLE IF EXISTS permissions;
DROP TABLE IF EXISTS roles;
DROP TABLE IF EXISTS users;
