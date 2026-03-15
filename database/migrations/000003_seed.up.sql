INSERT INTO users (name, email, age, gender, birth_date) VALUES
  ('Айгүл',    'aigul.m@gmail.com',       25, 'F', '1999-01-15'),
  ('Дмитрий',  'dmitry.k@mail.ru',        30, 'M', '1994-02-20'),
  ('Айнур',    'ainur92@yandex.ru',       28, 'F', '1996-03-10'),
  ('Александр','alex.s@mail.ru',          22, 'M', '2002-04-05'),
  ('Дана',     'dana.kz@gmail.com',       27, 'F', '1997-05-12'),
  ('Сергей',   'sergey.v@inbox.ru',      31, 'M', '1993-06-18'),
  ('Аружан',   'aruzhan@mail.ru',        24, 'F', '2000-07-22'),
  ('Максим',   'max.imov@gmail.com',      29, 'M', '1995-08-30'),
  ('Ольга',    'olga.p@yandex.ru',       26, 'F', '1998-09-14'),
  ('Нурлан',   'nurlan.b@mail.ru',        23, 'M', '2001-10-01'),
  ('Екатерина','katy.a@gmail.com',        32, 'F', '1992-11-11'),
  ('Ерлан',    'erlan.kz@yandex.ru',     21, 'M', '2003-12-25'),
  ('Мария',    'maria.s@mail.ru',         28, 'F', '1996-01-08'),
  ('Артём',    'artem.d@gmail.com',       30, 'M', '1994-02-14'),
  ('Жанар',    'zhanar.m@mail.ru',        25, 'F', '1999-03-03'),
  ('Иван',     'ivan.i@yandex.ru',        27, 'M', '1997-04-17'),
  ('Алия',     'aliya.s@gmail.com',        24, 'F', '2000-05-20'),
  ('Павел',    'pavel.v@mail.ru',         26, 'M', '1998-06-06'),
  ('Дарья',    'darya.k@inbox.ru',        29, 'F', '1995-07-09'),
  ('Бауыржан', 'bauyrzhan@yandex.ru',     31, 'M', '1993-08-12'),
  ('Анна',     'anna.n@gmail.com',        22, 'F', '2002-09-19'),
  ('Кайрат',   'kairat.t@mail.ru',       28, 'M', '1996-10-22'),
  ('Виктория', 'vika.l@yandex.ru',       23, 'F', '2001-11-30'),
  ('Тимур',    'timur.r@gmail.com',       27, 'M', '1997-12-07'),
  ('Сания',    'saniya.k@mail.ru',       26, 'F', '1998-01-11'),
  ('Андрей',   'andrey.m@inbox.ru',       24, 'M', '2000-02-28'),
  ('Айгерим',  'aigerim.z@gmail.com',     29, 'F', '1995-03-15')
ON CONFLICT (email) DO NOTHING;

INSERT INTO user_friends (user_id, friend_id)
SELECT u.id, f.id FROM users u, users f
WHERE u.email = 'aigul.m@gmail.com' AND f.email = 'dmitry.k@mail.ru'
ON CONFLICT (user_id, friend_id) DO NOTHING;
INSERT INTO user_friends (user_id, friend_id)
SELECT u.id, f.id FROM users u, users f
WHERE u.email = 'dmitry.k@mail.ru' AND f.email = 'aigul.m@gmail.com'
ON CONFLICT (user_id, friend_id) DO NOTHING;

INSERT INTO user_friends (user_id, friend_id)
SELECT u.id, f.id FROM users u, users f
WHERE u.email = 'aigul.m@gmail.com' AND f.email IN ('ainur92@yandex.ru','alex.s@mail.ru','dana.kz@gmail.com')
ON CONFLICT (user_id, friend_id) DO NOTHING;
INSERT INTO user_friends (user_id, friend_id)
SELECT f.id, u.id FROM users u, users f
WHERE u.email = 'aigul.m@gmail.com' AND f.email IN ('ainur92@yandex.ru','alex.s@mail.ru','dana.kz@gmail.com')
ON CONFLICT (user_id, friend_id) DO NOTHING;

INSERT INTO user_friends (user_id, friend_id)
SELECT u.id, f.id FROM users u, users f
WHERE u.email = 'dmitry.k@mail.ru' AND f.email IN ('ainur92@yandex.ru','alex.s@mail.ru','dana.kz@gmail.com')
ON CONFLICT (user_id, friend_id) DO NOTHING;
INSERT INTO user_friends (user_id, friend_id)
SELECT f.id, u.id FROM users u, users f
WHERE u.email = 'dmitry.k@mail.ru' AND f.email IN ('ainur92@yandex.ru','alex.s@mail.ru','dana.kz@gmail.com')
ON CONFLICT (user_id, friend_id) DO NOTHING;

INSERT INTO user_friends (user_id, friend_id)
SELECT u.id, f.id FROM users u, users f
WHERE u.email = 'ainur92@yandex.ru' AND f.email IN ('alex.s@mail.ru','dana.kz@gmail.com','sergey.v@inbox.ru')
ON CONFLICT (user_id, friend_id) DO NOTHING;
INSERT INTO user_friends (user_id, friend_id)
SELECT f.id, u.id FROM users u, users f
WHERE u.email = 'ainur92@yandex.ru' AND f.email IN ('alex.s@mail.ru','dana.kz@gmail.com','sergey.v@inbox.ru')
ON CONFLICT (user_id, friend_id) DO NOTHING;

INSERT INTO user_friends (user_id, friend_id)
SELECT u.id, f.id FROM users u, users f
WHERE u.email = 'alex.s@mail.ru' AND f.email IN ('dana.kz@gmail.com','sergey.v@inbox.ru')
ON CONFLICT (user_id, friend_id) DO NOTHING;
INSERT INTO user_friends (user_id, friend_id)
SELECT f.id, u.id FROM users u, users f
WHERE u.email = 'alex.s@mail.ru' AND f.email IN ('dana.kz@gmail.com','sergey.v@inbox.ru')
ON CONFLICT (user_id, friend_id) DO NOTHING;

INSERT INTO user_friends (user_id, friend_id)
SELECT u.id, f.id FROM users u, users f
WHERE u.email = 'dana.kz@gmail.com' AND f.email = 'sergey.v@inbox.ru'
ON CONFLICT (user_id, friend_id) DO NOTHING;
INSERT INTO user_friends (user_id, friend_id)
SELECT f.id, u.id FROM users u, users f
WHERE u.email = 'dana.kz@gmail.com' AND f.email = 'sergey.v@inbox.ru'
ON CONFLICT (user_id, friend_id) DO NOTHING;
