// Copyright (c) 2025 John Dewey

// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to
// deal in the Software without restriction, including without limitation the
// rights to use, copy, modify, merge, publish, distribute, sublicense, and/or
// sell copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:

// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.

// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING
// FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER
// DEALINGS IN THE SOFTWARE.

package cmd

import (
	"encoding/json"
	"log/slog"
	"time"

	"github.com/spf13/cobra"
)

// clientJobDeleteCmd represents the clientJobsDelete command.
var clientJobDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete a job from KV store",
	Long: `Deletes a specific job from the NATS KV store using its job ID.
This removes the job permanently from storage.`,
	Run: func(cmd *cobra.Command, _ []string) {
		jobID, _ := cmd.Flags().GetString("job-id")
		ctx := cmd.Context()

		err := jobClient.DeleteJob(ctx, jobID)
		if err != nil {
			logFatal("failed to delete job", err)
		}

		if jsonOutput {
			result := map[string]interface{}{
				"status":    "deleted",
				"job_id":    jobID,
				"timestamp": time.Now().Format(time.RFC3339),
			}
			resultJSON, _ := json.Marshal(result)
			logger.Info("job deleted", slog.String("response", string(resultJSON)))
			return
		}

		deleteData := map[string]interface{}{
			"Job ID": jobID,
			"Status": "Deleted",
		}
		printStyledMap(deleteData)

		logger.Info("job deleted successfully",
			slog.String("job_id", jobID),
		)
	},
}

func init() {
	clientJobCmd.AddCommand(clientJobDeleteCmd)

	clientJobDeleteCmd.PersistentFlags().
		StringP("job-id", "", "", "Job ID to delete")

	_ = clientJobDeleteCmd.MarkPersistentFlagRequired("job-id")
}
