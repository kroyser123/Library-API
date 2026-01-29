CREATE TABLE IF NOT EXISTS books (
 id VARCHAR(36) PRIMARY KEY,
 title VARCHAR(200) NOT NULL,
 author VARCHAR(200) NOT NULL,
 year INTEGER NOT NULL,
 created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
 updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_books_title ON books(title);
CREATE INDEX IF NOT EXISTS idx_books_author ON books(author);
CREATE INDEX IF NOT EXISTS idx_books_year ON books(year);

INSERT INTO books (id, title, author, year, created_at) VALUES
('550e8400-e29b-41d4-a716-446655440001', '1984', 'George Orwell', 1949, NOW()),
('550e8400-e29b-41d4-a716-446655440002', 'Animal Farm', 'George Orwell', 1945, NOW()),
('550e8400-e29b-41d4-a716-446655440003', 'Brave New World', 'Aldous Huxley', 1932, NOW()),
('550e8400-e29b-41d4-a716-446655440004', 'To Kill a Mockingbird', 'Harper Lee', 1960, NOW()),
('550e8400-e29b-41d4-a716-446655440005', 'The Great Gatsby', 'F. Scott Fitzgerald', 1925, NOW())
ON CONFLICT (id) DO NOTHING;