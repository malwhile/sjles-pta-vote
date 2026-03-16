import React, { useState, ChangeEvent, FormEvent, useEffect } from "react";
import { useNavigate } from 'react-router';
import {
  Card,
  TextField,
  Button,
  Alert,
  Box,
  Typography,
  TableContainer,
  Table,
  TableHead,
  TableBody,
  TableRow,
  TableCell,
  Paper,
} from "@mui/material";
import PageLayout from "../components/PageLayout";
import { Member } from "../types";

export default function AdminMembersView() {
  const [year, setYear] = useState<string>("");
  const [members, setMembers] = useState<Member[]>([]);
  const [status, setStatus] = useState<string>("");
  const [statusSeverity, setStatusSeverity] = useState<'success' | 'error'>('success');
  const navigate = useNavigate();

  useEffect(() => {
    const isAdmin = localStorage.getItem('adminToken') !== null;
    if (!isAdmin) {
      navigate('/admin-login');
    }
  }, [navigate]);

  const handleSubmit = async (e: FormEvent<HTMLFormElement>) => {
    e.preventDefault();

    try {
      const token = localStorage.getItem('adminToken');
      const headers: HeadersInit = {};
      if (token) {
        (headers as any)['Authorization'] = `Bearer ${token}`;
      }

      const resp = await fetch(`/api/admin/members/view?year=${year}`, {
        headers: headers,
      });
      const data = await resp.json();

      if (data.success) {
        setMembers(Array.isArray(data.data) ? data.data : []);
        setStatus("");
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
      <Card sx={{ p: 3, mb: 3, maxWidth: 500 }}>
        <Typography variant="h5" gutterBottom sx={{ fontWeight: 600, mb: 3 }}>
          View Members
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

          {status && (
            <Alert severity={statusSeverity} sx={{ my: 2 }}>
              {status}
            </Alert>
          )}

          <Button
            type="submit"
            variant="contained"
            fullWidth
            sx={{ mt: 2 }}
          >
            View Members
          </Button>
        </Box>
      </Card>

      {members.length > 0 && (
        <TableContainer component={Paper}>
          <Table>
            <TableHead>
              <TableRow>
                <TableCell sx={{ fontWeight: 600 }}>Name</TableCell>
                <TableCell sx={{ fontWeight: 600 }}>Email</TableCell>
              </TableRow>
            </TableHead>
            <TableBody>
              {members.map((member, index) => (
                <TableRow key={index} hover>
                  <TableCell>{member.Name}</TableCell>
                  <TableCell>{member.Email}</TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </TableContainer>
      )}
    </PageLayout>
  );
}
