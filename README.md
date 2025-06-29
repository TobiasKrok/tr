# tr - translation tool for spanish to english and vice versa

vibe coded and WIP cause I wanted it

## Build

```bash
go build -o tr.exe ./cmd/tr
```

## Usage

### Interactive Mode

```bash
./tr.exe
```

- Type words/phrases and press Enter to translate
- Use `Ctrl+T` to toggle direction (ES→EN or EN→ES)
- Type `exit` or use `Ctrl+C` to quit

#### Examples

```
> hola
┌─────────┬─────────┐
│ Spanish │ English │
├─────────┼─────────┤
│ hola    │ hello   │
└─────────┴─────────┘

> caminar
┌─────────┬─────────┐
│ Spanish │ English │
├─────────┼─────────┤
│ caminar │ to walk │
└─────────┴─────────┘

Verb Conjugations:
┌──────────┬─────────┬───────────┬───────────┐
│ Person   │ Present │ Preterite │ Imperfect │
├──────────┼─────────┼───────────┼───────────┤
│ yo       │ camino  │ caminé    │ caminaba  │
│ tú       │ caminas │ caminaste │ caminabas │
│ él/ella  │ camina  │ caminó    │ caminaba  │
│ nosotros │ caminamos│ caminamos │ caminábamos│
│ vosotros │ camináis│ caminasteis│ caminabais│
│ ellos    │ caminan │ caminaron │ caminaban │
└──────────┴─────────┴───────────┴───────────┘
```

### Command Line

```bash
# Spanish to English (default)
./tr.exe hola
# Output: hello

./tr.exe "buenos días"
# Output: good morning

# English to Spanish
./tr.exe -d en2es hello
# Output: hola

./tr.exe -d en2es "good morning"
# Output: buenos días

# Explicit Spanish to English
./tr.exe -d es2en caminar
# Output: to walk (+ conjugation table for verbs)
```

### Options

- `-d, --direction`: `es2en` (default) or `en2es`
- `-h, --help`: Show help
- `-v, --version`: Show version

Verb conjugations are automatically shown for Spanish verbs.



