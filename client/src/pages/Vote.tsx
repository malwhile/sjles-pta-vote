import React, { useState, useEffect } from "react";
import { useParams, useNavigate } from "react-router";
import axios from "axios";
import {
  Typography,
  Box,
  Card,
  TextField,
  Button,
  CircularProgress,
  Alert,
} from "@mui/material";
import PageLayout from "../components/PageLayout";
import { Poll } from "../types";

export default function Vote() {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const [poll, setPoll] = useState<Poll | null>(null);
  const [loading, setLoading] = useState(true);
  const [email, setEmail] = useState("");
  const [selectedVote, setSelectedVote] = useState<boolean | null>(null);
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [successMessage, setSuccessMessage] = useState<string | null>(null);

  useEffect(() => {
    if (id) fetchPollDetails(parseInt(id, 10));
  }, [id]);

  const fetchPollDetails = async (pollId: number) => {
    try {
      setLoading(true);
      const response = await axios.post(`/api/admin/view-polls`, { poll_id: pollId });
      setPoll(response.data);
    } catch (e) {
      console.error("Error fetching poll details:", e);
      setError("Failed to load poll");
    } finally {
      setLoading(false);
    }
  };

  const handleSubmitVote = async () => {
    if (!email || selectedVote === null || !poll) {
      setError("Please enter an email and select Yes or No");
      return;
    }

    setSubmitting(true);
    setError(null);

    try {
      const response = await axios.post("/api/vote", {
        poll_id: poll.id,
        email,
        vote: selectedVote,
      });

      if (response.status === 200) {
        setSuccessMessage("Vote submitted successfully!");
        setTimeout(() => {
          navigate(`/poll-details/${poll.id}`);
        }, 500);
      }
    } catch (e: any) {
      console.error("Error submitting vote:", e);
      if (e.response?.status === 409) {
        setError("You have already voted on this poll.");
      } else {
        setError("Failed to submit vote. Please try again.");
      }
    } finally {
      setSubmitting(false);
    }
  };

  if (loading) {
    return (
      <PageLayout>
        <Box display="flex" justifyContent="center" alignItems="center" minHeight="400px">
          <CircularProgress />
        </Box>
      </PageLayout>
    );
  }

  if (!poll) {
    return (
      <PageLayout>
        <Typography>Poll not found</Typography>
      </PageLayout>
    );
  }

  return (
    <PageLayout>
      <Card sx={{ p: 3, maxWidth: 500, mx: "auto" }}>
        <Typography variant="h5" gutterBottom sx={{ fontWeight: 600, mb: 2 }}>
          {poll.question}
        </Typography>

        <Typography variant="body2" color="textSecondary" sx={{ mb: 3 }}>
          Your vote is anonymous — your email is only used to prevent duplicate votes.
        </Typography>

        {error && (
          <Alert severity="error" sx={{ mb: 2 }}>
            {error}
          </Alert>
        )}

        {successMessage && (
          <Alert severity="success" sx={{ mb: 2 }}>
            {successMessage}
          </Alert>
        )}

        <Box sx={{ mb: 3 }}>
          <TextField
            fullWidth
            type="email"
            label="Email"
            placeholder="your@email.com"
            value={email}
            onChange={(e) => setEmail(e.target.value)}
            disabled={submitting}
            margin="normal"
          />
        </Box>

        <Box sx={{ display: "flex", gap: 2, mb: 3 }}>
          <Button
            fullWidth
            variant={selectedVote === true ? "contained" : "outlined"}
            color="success"
            onClick={() => setSelectedVote(true)}
            disabled={submitting}
          >
            Yes
          </Button>
          <Button
            fullWidth
            variant={selectedVote === false ? "contained" : "outlined"}
            color="error"
            onClick={() => setSelectedVote(false)}
            disabled={submitting}
          >
            No
          </Button>
        </Box>

        <Button
          fullWidth
          variant="contained"
          onClick={handleSubmitVote}
          disabled={submitting || !email || selectedVote === null}
        >
          {submitting ? "Submitting..." : "Submit Vote"}
        </Button>
      </Card>
    </PageLayout>
  );
}
