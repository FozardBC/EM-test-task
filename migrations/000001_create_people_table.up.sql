CREATE TABLE people (
    id SERIAL PRIMARY KEY,
    name VARCHAR(50) NOT NULL,
    surname VARCHAR(50) NOT NULL,
    patronymic VARCHAR(50) NOT NULL,
    age INTEGER,
    gender VARCHAR(4),
    nationality VARCHAR(2),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_people_name ON people USING HASH (name);
CREATE INDEX idx_people_surname ON people USING HASH (surname);
CREATE INDEX idx_people_age ON people(age);
CREATE INDEX idx_people_gender ON people(gender);
CREATE INDEX idx_people_nationality ON people(nationality);