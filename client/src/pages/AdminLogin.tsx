import React, { useState, ChangeEvent, FormEvent } from "react";
import { useNavigate } from "react-router";
import {
  Card,
  TextField,
  Button,
  Alert,
  Box,
} from "@mui/material";
import PageLayout from "../components/PageLayout";

export default function AdminLogin() {
  const [username, setUsername] = useState<string>("");
  const [password, setPassword] = useState<string>("");
  const [error, setError] = useState<string>("");
  const [isLoading, setIsLoading] = useState<boolean>(false);
  const navigate = useNavigate();

  const handleSubmit = async (e: FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    setError("");
    setIsLoading(true);

    if (!username || !password) {
      setError("Please enter both username and password.");
      setIsLoading(false);
      return;
    }

    try {
      const resp = await fetch("/api/admin/login", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify({
          username,
          password,
        }),
      });

      const data = await resp.json();

      if (resp.ok && data.success) {
        localStorage.setItem("adminToken", data.token);
        setError("");
        navigate("/admin-members");
      } else {
        setError(data.error || "Login failed");
      }
    } catch (err) {
      const errorMsg = err instanceof Error ? err.message : String(err);
      setError(`Network error: ${errorMsg}`);
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <PageLayout maxWidth="sm">
      <Card sx={{ mt: 8, p: 3 }}>
        <Box component="form" onSubmit={handleSubmit}>
          <TextField
            fullWidth
            label="Username"
            type="text"
            value={username}
            onChange={(e: ChangeEvent<HTMLInputElement>) => setUsername(e.target.value)}
            margin="normal"
            placeholder="Enter your username"
          />

          <TextField
            fullWidth
            label="Password"
            type="password"
            value={password}
            onChange={(e: ChangeEvent<HTMLInputElement>) => setPassword(e.target.value)}
            margin="normal"
            placeholder="Enter your password"
          />

          {error && (
            <Alert severity="error" sx={{ mt: 2, mb: 2 }}>
              {error}
            </Alert>
          )}

          <Button
            type="submit"
            variant="contained"
            fullWidth
            disabled={isLoading}
            sx={{ mt: 3 }}
          >
            {isLoading ? "Logging in..." : "Login"}
          </Button>
        </Box>
      </Card>
    </PageLayout>
  );
}
