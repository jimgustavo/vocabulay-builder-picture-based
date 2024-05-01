CREATE TABLE data_set (
    id SERIAL PRIMARY KEY,
    category TEXT NOT NULL,
    question TEXT NOT NULL,
    targetWord TEXT NOT NULL,
    answers JSONB NOT NULL,
    correct INTEGER NOT NULL
);
