import React, { useState, ChangeEvent, FormEvent } from 'react';
import { useNavigate } from 'react-router';
import {
  Card,
  TextField,
  Button,
  Alert,
  Box,
  Typography,
} from "@mui/material";
import PageLayout from "../components/PageLayout";

export default function AdminCreateVote() {
  const navigate = useNavigate();
  const [question, setQuestion] = useState<string>('');
  const [expiresInHours, setExpiresInHours] = useState<string>('');
  const [status, setStatus] = useState<string>("");
  const [statusSeverity, setStatusSeverity] = useState<'success' | 'error'>('success');

  const isAdmin = () => {
    return localStorage.getItem('adminToken') !== null;
  };

  if (!isAdmin()) {
    navigate('/admin-login');
    return <div>Redirecting...</div>;
  }

  const handleSubmit = async (e: FormEvent<HTMLFormElement>) => {
    e.preventDefault();

    const formData = new FormData();
    formData.append("question", question);
    formData.append("duration", expiresInHours);

    try {
      const resp = await fetch("/api/admin/new-poll", {
        method: "POST",
        body: formData,
      });

      const data = await resp.json();

      if (data.success) {
        setStatus("Vote created successfully!");
        setStatusSeverity('success');
        setQuestion("");
        setExpiresInHours("");
      } else {
        setStatus(data.error || "Server error");
        setStatusSeverity('error');
      }
    } catch (error) {
      const errorMsg = error instanceof Error ? error.message : String(error);
      setStatus(`Failed to create vote: ${errorMsg}`);
      setStatusSeverity('error');
    }
  };

  return (
    <PageLayout>
      <Card sx={{ p: 3, maxWidth: 500 }}>
        <Typography variant="h5" gutterBottom sx={{ fontWeight: 600, mb: 3 }}>
          Create New Vote
        </Typography>

        <Box component="form" onSubmit={handleSubmit}>
          <TextField
            fullWidth
            label="Question"
            type="text"
            value={question}
            onChange={(e: ChangeEvent<HTMLInputElement>) => setQuestion(e.target.value)}
            margin="normal"
            required
          />

          <TextField
            fullWidth
            label="Expires In (hours)"
            type="number"
            value={expiresInHours}
            onChange={(e: ChangeEvent<HTMLInputElement>) => setExpiresInHours(e.target.value)}
            margin="normal"
            inputProps={{ min: 1 }}
            required
          />

          {status && (
            <Alert severity={statusSeverity} sx={{ my: 2 }}>
              {status}
            </Alert>
          )}

          <Button
            type="submit"
            variant="contained"
            fullWidth
            sx={{ mt: 3 }}
          >
            Create Vote
          </Button>
        </Box>
      </Card>
    </PageLayout>
  );
}
