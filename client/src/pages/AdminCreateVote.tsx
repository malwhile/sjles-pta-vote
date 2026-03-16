import React, { useState, ChangeEvent, FormEvent, useEffect } from 'react';
import { useNavigate } from 'react-router';
import axios from 'axios';
import {
  Card,
  TextField,
  Button,
  Alert,
  Box,
  Typography,
} from "@mui/material";
import PageLayout from "../components/PageLayout";
import { getAuthHeaders } from '../utils/api';

export default function AdminCreateVote() {
  const navigate = useNavigate();
  const [question, setQuestion] = useState<string>('');
  const [expiresInHours, setExpiresInHours] = useState<string>('');
  const [status, setStatus] = useState<string>("");
  const [statusSeverity, setStatusSeverity] = useState<'success' | 'error'>('success');

  useEffect(() => {
    const isAdmin = localStorage.getItem('adminToken') !== null;
    if (!isAdmin) {
      navigate('/admin-login');
    }
  }, [navigate]);

  const handleSubmit = async (e: FormEvent<HTMLFormElement>) => {
    e.preventDefault();

    try {
      const authHeaders = getAuthHeaders();
      const resp = await axios.post("/api/admin/new-poll", {
        question,
        duration_hours: parseInt(expiresInHours, 10),
      }, authHeaders);

      if (resp.data.success || resp.status === 200) {
        setStatus("Vote created successfully!");
        setStatusSeverity('success');
        setQuestion("");
        setExpiresInHours("");
      } else {
        setStatus(resp.data.error || "Server error");
        setStatusSeverity('error');
      }
    } catch (error: any) {
      let errorMsg = "Failed to create vote. Please try again.";
      if (error.response?.status === 400) {
        errorMsg = error.response?.data?.error || "Invalid vote details. Please check your input.";
      } else if (error.response?.status === 401) {
        errorMsg = "Your session has expired. Please log in again.";
        localStorage.removeItem('adminToken');
        setTimeout(() => navigate('/admin-login'), 2000);
      } else if (error.response?.status === 403) {
        errorMsg = "You are not authorized to create votes.";
      }
      setStatus(errorMsg);
      setStatusSeverity('error');
      if (process.env.NODE_ENV === "development") {
        console.error("Error creating vote:", error);
      }
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
