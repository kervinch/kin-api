CREATE TYPE gender_enum AS ENUM ('male', 'female', '');
ALTER TABLE users ADD COLUMN gender gender_enum;
ALTER TABLE users ADD COLUMN date_of_birth timestamp(0) with time zone NULL;
ALTER TABLE users ADD COLUMN phone_number text;