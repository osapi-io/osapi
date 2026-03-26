// Copyright (c) 2024 John Dewey

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

package config

// AllKVBuckets returns all KV bucket configurations declared in the NATS
// config. Each entry carries a human-readable name and the bucket name from
// config. Entries with an empty Bucket field are still included so that
// callers can filter by their own policy.
func (n NATS) AllKVBuckets() []KVBucketInfo {
	return []KVBucketInfo{
		{Name: "job-queue", Bucket: n.KV.Bucket},
		{Name: "job-responses", Bucket: n.KV.ResponseBucket},
		{Name: "audit", Bucket: n.Audit.Bucket},
		{Name: "registry", Bucket: n.Registry.Bucket},
		{Name: "facts", Bucket: n.Facts.Bucket},
		{Name: "state", Bucket: n.State.Bucket},
		{Name: "file-state", Bucket: n.FileState.Bucket},
	}
}

// AllObjectStoreBuckets returns all Object Store bucket configurations
// declared in the NATS config. Entries with an empty Bucket field are still
// included so that callers can filter by their own policy.
func (n NATS) AllObjectStoreBuckets() []ObjectStoreBucketInfo {
	return []ObjectStoreBucketInfo{
		{Name: "file-objects", Bucket: n.Objects.Bucket},
	}
}
