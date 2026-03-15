DELETE FROM user_friends;
DELETE FROM users WHERE email LIKE '%@example.com' OR email IN (
  'aigul.m@gmail.com','dmitry.k@mail.ru','ainur92@yandex.ru','alex.s@mail.ru','dana.kz@gmail.com',
  'sergey.v@inbox.ru','aruzhan@mail.ru','max.imov@gmail.com','olga.p@yandex.ru','nurlan.b@mail.ru',
  'katy.a@gmail.com','erlan.kz@yandex.ru','maria.s@mail.ru','artem.d@gmail.com','zhanar.m@mail.ru',
  'ivan.i@yandex.ru','aliya.s@gmail.com','pavel.v@mail.ru','darya.k@inbox.ru','bauyrzhan@yandex.ru',
  'anna.n@gmail.com','kairat.t@mail.ru','vika.l@yandex.ru','timur.r@gmail.com','saniya.k@mail.ru',
  'andrey.m@inbox.ru','aigerim.z@gmail.com'
);
