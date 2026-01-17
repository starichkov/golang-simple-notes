# API Reference

The Notes API provides both REST and gRPC interfaces for managing notes.

## 📖 API Reference

### REST API

- `GET /api/notes` - List all notes
- `GET /api/notes/{id}` - Get a note by ID
- `POST /api/notes` - Create a new note
- `PUT /api/notes/{id}` - Update a note
- `DELETE /api/notes/{id}` - Delete a note

#### Example Request (Create Note)
```bash
curl -X POST http://localhost:8080/api/notes \
  -H "Content-Type: application/json" \
  -d '{"title":"My Note","content":"This is the content of my note"}'
```

### gRPC API

Service: `notes.Notes`
- `CreateNote`: Create a new note
- `GetNote`: Get a note by ID
- `GetAllNotes`: Get all notes
- `UpdateNote`: Update an existing note
- `DeleteNote`: Delete a note
