-- The interface now defaults to the brand accent taken from the logo, rather
-- than shadcn's near-black neutral.
--
-- 'neutral' was the column default since 017, so a stored 'neutral' means "never
-- chose" far more often than "chose grey". Those rows move to the new default;
-- anyone who actually wants the greyscale interface can still select Neutral,
-- which is now an explicit theme rather than the absence of one.
ALTER TABLE user_preference ALTER COLUMN color_theme SET DEFAULT 'brand';
UPDATE user_preference SET color_theme = 'brand' WHERE color_theme = 'neutral';
