# Multilingual Test Data for Anonymization

This directory contains test documents in multiple European languages for testing the anonymization pipeline.

## Files

- **german.txt** - German text with named entities (persons, organizations, locations)
- **french.txt** - French text with named entities
- **english.txt** - English text with named entities
- **spanish.txt** - Spanish text with named entities

## Entity Types Present

Each file contains examples of:
- **PERSON**: Names of individuals (e.g., Angela Merkel, Emmanuel Macron, Joe Biden)
- **ORG**: Organizations and companies (e.g., Siemens AG, TotalEnergies, Apple Inc.)
- **LOC**: Locations (e.g., Berlin, Bruxelles, White House)
- **MISC**: Other entities (roles, events, etc.)

## Coreference Examples

The texts include coreference chains:
- "Angela Merkel ... die ehemalige Bundeskanzlerin ... Sie"
- "Joe Biden ... The President ... he"
- "Emmanuel Macron ... le Président ... Il"

## Use Cases

1. **NER Testing**: Extract entities from multilingual text
2. **Coref Testing**: Resolve mentions to same entities
3. **Synonym Testing**: Link variants (e.g., "Merkel" and "Angela Merkel")
4. **Anonymization Testing**: Replace entities with pseudonyms consistently
5. **Reversibility Testing**: Map pseudonyms back to original names

## Expected Output Example

Input (German):
```
Angela Merkel besuchte Berlin. Die Bundeskanzlerin traf Franziska Giffey.
```

Anonymized:
```
PERSON_001234 besuchte LOC_005678. Die Bundeskanzlerin traf PERSON_009876.
```

With mappings:
- PERSON_001234 ↔ Angela Merkel
- LOC_005678 ↔ Berlin
- PERSON_009876 ↔ Franziska Giffey

## Adding More Languages

To add test data for other languages:

1. Create `{language}.txt` file
2. Include 10-15 named entities of various types
3. Add coreference chains (pronouns, titles referring to same person)
4. Include company names, locations, and people
5. Ensure realistic context (news article style works well)

Languages to consider adding:
- Italian
- Portuguese
- Polish
- Dutch
- Swedish
- Czech

