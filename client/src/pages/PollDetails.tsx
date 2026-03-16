import React, { useState, useEffect } from "react";
import { useParams } from "react-router";
import axios from "axios";
import {
  Typography,
  Box,
  Card,
  Chip,
  CircularProgress,
  Link,
  Alert,
} from "@mui/material";
import { PieChart, Pie, Tooltip, Legend, Cell, ResponsiveContainer } from "recharts";
import QRCode from "react-qr-code";
import PageLayout from "../components/PageLayout";
import { Poll } from "../types";

const COLORS = ["#0088FE", "#FEB43C"];

export default function PollDetails() {
  const { id } = useParams<{ id: string }>();
  const [poll, setPoll] = useState<Poll | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (id) fetchPollDetails(parseInt(id, 10));
  }, [id]);

  const fetchPollDetails = async (pollId: number) => {
    try {
      setLoading(true);
      setError(null);
      const response = await axios.post(`/api/admin/view-polls`, { poll_id: pollId });
      setPoll(response.data);
    } catch (e: any) {
      const errorMsg = e.response?.status === 404 ? "Poll not found" : "Failed to load poll. Please try again.";
      setError(errorMsg);
      if (process.env.NODE_ENV === "development") {
        console.error("Error fetching poll details:", e);
      }
    } finally {
      setLoading(false);
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

  if (error) {
    return (
      <PageLayout>
        <Card sx={{ p: 3 }}>
          <Alert severity="error">{error}</Alert>
        </Card>
      </PageLayout>
    );
  }

  if (!poll) {
    return (
      <PageLayout>
        <Card sx={{ p: 3 }}>
          <Alert severity="warning">Poll not found</Alert>
        </Card>
      </PageLayout>
    );
  }

  const totalMemberVotes = poll.member_yes + poll.member_no;
  const hasVotes = totalMemberVotes > 0;
  const passed = hasVotes && poll.member_yes > poll.member_no;

  const memberData = [
    { name: "Yes", value: poll.member_yes },
    { name: "No", value: poll.member_no },
  ];

  const nonMemberData = [
    { name: "Yes", value: poll.non_member_yes },
    { name: "No", value: poll.non_member_no },
  ];

  const voteUrl = `${window.location.origin}/vote/${poll.id}`;

  return (
    <PageLayout>
      <Box sx={{ mb: 3 }}>
        <Box sx={{ display: "flex", alignItems: "center", gap: 2, mb: 2 }}>
          <Typography variant="h4" sx={{ fontWeight: 600 }}>
            {poll.question}
          </Typography>
          <Chip
            label={!hasVotes ? "No votes yet" : (passed ? "Pass" : "Fail")}
            color={!hasVotes ? "default" : (passed ? "success" : "error")}
            size="medium"
          />
        </Box>
        <Typography variant="body2" color="textSecondary">
          Created: {new Date(poll.created_at).toLocaleString()}
        </Typography>
      </Box>

      <Card sx={{ p: 3, mb: 3, display: "flex", flexDirection: "column", alignItems: "center", gap: 2 }}>
        <Typography variant="h6" sx={{ fontWeight: 600 }}>
          Scan to Vote
        </Typography>
        <QRCode value={voteUrl} size={200} />
        <Link href={voteUrl} target="_blank" rel="noopener noreferrer" sx={{ fontSize: 12, wordBreak: "break-all", textAlign: "center" }}>
          {voteUrl}
        </Link>
      </Card>

      <Box sx={{ display: { xs: 'block', md: 'flex' }, gap: 3 }}>
        <Card sx={{ p: 2, flex: 1, minWidth: { md: 0 } }}>
          <Typography variant="h6" gutterBottom sx={{ fontWeight: 600 }}>
            Member Votes
          </Typography>
          <ResponsiveContainer width="100%" height={280}>
            <PieChart>
              <Pie
                data={memberData}
                cx="50%"
                cy="50%"
                labelLine={false}
                outerRadius={100}
                fill="#8884d8"
                dataKey="value"
              >
                {memberData.map((_, i) => (
                  <Cell key={`cell-${i}`} fill={COLORS[i % COLORS.length]} />
                ))}
              </Pie>
              <Tooltip />
            </PieChart>
          </ResponsiveContainer>
          <Box sx={{ mt: 2, display: "flex", justifyContent: "center", gap: 2 }}>
            <Box sx={{ display: "flex", alignItems: "center", gap: 1 }}>
              <Box sx={{ width: 16, height: 16, backgroundColor: COLORS[0] }} />
              <Typography variant="body2">Yes: {poll.member_yes}</Typography>
            </Box>
            <Box sx={{ display: "flex", alignItems: "center", gap: 1 }}>
              <Box sx={{ width: 16, height: 16, backgroundColor: COLORS[1] }} />
              <Typography variant="body2">No: {poll.member_no}</Typography>
            </Box>
          </Box>
        </Card>

        <Card sx={{ p: 2, flex: 1, minWidth: { md: 0 }, mt: { xs: 3, md: 0 } }}>
          <Typography variant="h6" gutterBottom sx={{ fontWeight: 600 }}>
            Non-Member Votes
          </Typography>
          <ResponsiveContainer width="100%" height={280}>
            <PieChart>
              <Pie
                data={nonMemberData}
                cx="50%"
                cy="50%"
                labelLine={false}
                outerRadius={100}
                fill="#8884d8"
                dataKey="value"
              >
                {nonMemberData.map((_, i) => (
                  <Cell key={`cell-${i}`} fill={COLORS[i % COLORS.length]} />
                ))}
              </Pie>
              <Tooltip />
            </PieChart>
          </ResponsiveContainer>
          <Box sx={{ mt: 2, display: "flex", justifyContent: "center", gap: 2 }}>
            <Box sx={{ display: "flex", alignItems: "center", gap: 1 }}>
              <Box sx={{ width: 16, height: 16, backgroundColor: COLORS[0] }} />
              <Typography variant="body2">Yes: {poll.non_member_yes}</Typography>
            </Box>
            <Box sx={{ display: "flex", alignItems: "center", gap: 1 }}>
              <Box sx={{ width: 16, height: 16, backgroundColor: COLORS[1] }} />
              <Typography variant="body2">No: {poll.non_member_no}</Typography>
            </Box>
          </Box>
        </Card>
      </Box>
    </PageLayout>
  );
}
