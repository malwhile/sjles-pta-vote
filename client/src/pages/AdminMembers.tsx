import React, { useState, ChangeEvent, FormEvent } from "react";
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

export default function AdminMembers() {
  const [year, setYear] = useState<string>("");
  const [file, setFile] = useState<File | null>(null);
  const [status, setStatus] = useState<string>("");
  const [statusSeverity, setStatusSeverity] = useState<'success' | 'error'>('success');

  const navigate = useNavigate();

  const isAdmin = () => {
    return localStorage.getItem('adminToken') !== null;
  };

  if (!isAdmin()) {
    navigate('/admin-login');
    return <div>Redirecting...</div>;
  }

  const handleSubmit = async (e: FormEvent<HTMLFormElement>) => {
    e.preventDefault();

    if (!year) {
      setStatus("Please enter a year.");
      setStatusSeverity('error');
      return;
    }
    if (!file) {
      setStatus("Please select a CSV file.");
      setStatusSeverity('error');
      return;
    }

    const formData = new FormData();
    formData.append("year", year);
    formData.append("members.csv", file);

    try {
      const token = localStorage.getItem('adminToken');
      const headers: HeadersInit = {};
      if (token) {
        (headers as any)['Authorization'] = `Bearer ${token}`;
      }

      const resp = await fetch("/api/admin/members", {
        method: "POST",
        headers: headers,
        body: formData,
      });

      const data = await resp.json();

      if (data.success) {
        setStatus("Uploaded successfully!");
        setStatusSeverity('success');
        setYear("");
        setFile(null);
      } else if (resp.status === 401) {
        setStatus("Your session has expired. Please log in again.");
        setStatusSeverity('error');
        localStorage.removeItem('adminToken');
        setTimeout(() => navigate('/admin-login'), 2000);
      } else {
        setStatus(data.error || "Server error");
        setStatusSeverity('error');
      }
    } catch (err) {
      const errorMsg = err instanceof Error ? err.message : String(err);
      setStatus(`Network error: ${errorMsg}`);
      setStatusSeverity('error');
    }
  };

  return (
    <PageLayout>
      <Card sx={{ p: 3, maxWidth: 500 }}>
        <Typography variant="h5" gutterBottom sx={{ fontWeight: 600, mb: 3 }}>
          Upload Members CSV
        </Typography>

        <Box component="form" onSubmit={handleSubmit}>
          <TextField
            fullWidth
            label="Year"
            type="number"
            value={year}
            onChange={(e: ChangeEvent<HTMLInputElement>) => setYear(e.target.value)}
            margin="normal"
            inputProps={{ min: 1900, max: 2100 }}
          />

          <Box sx={{ my: 2 }}>
            <Button
              variant="outlined"
              component="label"
              fullWidth
              sx={{ justifyContent: 'flex-start', textAlign: 'left' }}
            >
              {file ? file.name : 'Choose CSV File'}
              <input
                hidden
                type="file"
                accept=".csv"
                onChange={(e: ChangeEvent<HTMLInputElement>) => setFile(e.target.files?.[0] || null)}
              />
            </Button>
          </Box>

          {status && (
            <Alert severity={statusSeverity} sx={{ mb: 2 }}>
              {status}
            </Alert>
          )}

          <Button
            type="submit"
            variant="contained"
            fullWidth
            sx={{ mt: 2 }}
          >
            Upload
          </Button>
        </Box>
      </Card>
    </PageLayout>
  );
}
