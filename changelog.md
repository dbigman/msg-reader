## [Unreleased]
### Added
- **Email Extraction Notebook:** A new Jupyter Notebook that:
  - Extracts key details (subject, sender, recipients, body) from `.msg` email files.
  - Saves the email content as a `.txt` file.
  - Extracts attachments to an `attachments` folder (with unique naming for duplicates).
  - Combines all individual `.txt` files into one consolidated file.
