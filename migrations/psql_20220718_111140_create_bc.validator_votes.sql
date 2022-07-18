
-- +migrate Up
CREATE TABLE IF NOT EXISTS bc.validator_votes(
    id uuid NOT NULL PRIMARY KEY,
    wallet_id UUID  NOT NULL,
    lottery_id UUID  NOT NULL,
    participants_id UUID  NOT NULL,
    hash VARCHAR (255) NOT NULL,
    vote BOOLEAN  NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT now(),
    updated_at TIMESTAMP NOT NULL DEFAULT now()
);

ALTER TABLE bc.validator_votes ADD CONSTRAINT fk_validator_votes_lottery FOREIGN KEY (lottery_id) REFERENCES bc.lottery_table(id);
ALTER TABLE bc.validator_votes ADD CONSTRAINT fk_validator_votes_participants FOREIGN KEY (participants_id) REFERENCES bc.participants_table(id);
ALTER TABLE bc.validator_votes ADD CONSTRAINT fk_validator_votes_wallets FOREIGN KEY (wallet_id) REFERENCES auth.wallets(id);

-- +migrate Down
DROP TABLE IF EXISTS bc.validator_votes;
