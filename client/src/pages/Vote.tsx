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
  const [voteSubmitted, setVoteSubmitted] = useState(false);
  const [isExpired, setIsExpired] = useState(false);

  useEffect(() => {
    if (id) fetchPollDetails(parseInt(id, 10));
  }, [id]);

  const fetchPollDetails = async (pollId: number) => {
    try {
      setLoading(true);
      setError(null);
      setIsExpired(false);
      // Use public endpoint - no authentication required
      const response = await axios.get(`/api/polls/${pollId}`);
      // API returns {success: true, data: poll}
      const pollData = response.data.data;
      setPoll(pollData);

      // Check if poll has expired
      if (pollData.expires_at) {
        const expiryTime = new Date(pollData.expires_at);
        if (new Date() > expiryTime) {
          setIsExpired(true);
        }
      }
    } catch (e: any) {
      const errorMsg = e.response?.status === 404 ? "Poll not found" : "Failed to load poll";
      setError(errorMsg);
      if (process.env.NODE_ENV === "development") {
        console.error("Error fetching poll details:", e);
      }
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
        setVoteSubmitted(true);
      }
    } catch (e: any) {
      if (process.env.NODE_ENV === "development") {
        console.error("Error submitting vote:", e);
      }
      if (e.response?.status === 409) {
        setError("You have already voted on this poll.");
      } else if (e.response?.status === 403) {
        setError("This poll has expired and is no longer accepting votes.");
        setIsExpired(true);
      } else if (e.response?.status === 400) {
        setError("Invalid vote submission. Please check your input.");
      } else if (e.response?.status === 404) {
        setError("Poll not found.");
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

        {!voteSubmitted ? (
          <>
            <Typography variant="body2" color="textSecondary" sx={{ mb: 3 }}>
              Your vote is anonymous — your email is only used to prevent duplicate votes.
            </Typography>

            {isExpired && (
              <Alert severity="warning" sx={{ mb: 2 }}>
                This poll has expired and is no longer accepting votes.
              </Alert>
            )}

            {error && (
              <Alert severity="error" sx={{ mb: 2 }}>
                {error}
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
                disabled={submitting || isExpired}
              >
                Yes
              </Button>
              <Button
                fullWidth
                variant={selectedVote === false ? "contained" : "outlined"}
                color="error"
                onClick={() => setSelectedVote(false)}
                disabled={submitting || isExpired}
              >
                No
              </Button>
            </Box>

            <Button
              fullWidth
              variant="contained"
              onClick={handleSubmitVote}
              disabled={submitting || !email || selectedVote === null || isExpired}
            >
              {submitting ? "Submitting..." : "Submit Vote"}
            </Button>
          </>
        ) : (
          <>
            <Alert severity="success" sx={{ mb: 3 }}>
              Thank you! Your vote has been recorded.
            </Alert>

            <Box sx={{ mb: 3, textAlign: "center" }}>
              <Typography variant="h6" sx={{ fontWeight: 600, mb: 2 }}>
                You voted: <strong>{selectedVote ? "Yes" : "No"}</strong>
              </Typography>
            </Box>

            <Button
              fullWidth
              variant="outlined"
              onClick={() => navigate(`/poll-details/${poll.id}`)}
            >
              View Poll Results
            </Button>
          </>
        )}

        {error && voteSubmitted && (
          <Alert severity="error" sx={{ mt: 2 }}>
            {error}
          </Alert>
        )}
      </Card>
    </PageLayout>
  );
}
