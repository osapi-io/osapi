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

import "fmt"

// AllKVBuckets returns all KV bucket configurations declared in the NATS
// config. Each entry carries a human-readable name and the bucket name from
// config. Entries with an empty Bucket field are still included so that
// callers can filter by their own policy.
func (n NATS) AllKVBuckets() []KVBucketInfo {
	return []KVBucketInfo{
		{Name: "job-queue", Key: "nats.kv.bucket", Bucket: n.KV.Bucket},
		{Name: "registry", Key: "nats.registry.bucket", Bucket: n.Registry.Bucket},
		{Name: "facts", Key: "nats.facts.bucket", Bucket: n.Facts.Bucket},
		{Name: "state", Key: "nats.state.bucket", Bucket: n.State.Bucket},
		{Name: "file-state", Key: "nats.file_state.bucket", Bucket: n.FileState.Bucket},
		{Name: "enrollment", Key: "nats.enrollment.bucket", Bucket: n.Enrollment.Bucket},
		{Name: "secrets", Key: "nats.secrets.bucket", Bucket: n.Secrets.Bucket},
	}
}

// AllObjectStoreBuckets returns all Object Store bucket configurations
// declared in the NATS config. Entries with an empty Bucket field are still
// included so that callers can filter by their own policy.
func (n NATS) AllObjectStoreBuckets() []ObjectStoreBucketInfo {
	return []ObjectStoreBucketInfo{
		{Name: "file-objects", Key: "nats.objects.bucket", Bucket: n.Objects.Bucket},
	}
}

// sharedBucket returns an error when two settings name the same bucket. Each
// role keeps its own keys, and a shared bucket mixes them: the job queue
// would list registrations as jobs, for example. An unset bucket is a role
// that is off, and is not compared. An Object Store bucket is compared only
// with the other Object Store buckets, because NATS keeps the two kinds in
// different streams.
func (n NATS) sharedBucket() error {
	kv := n.AllKVBuckets()
	keys, buckets := make([]string, 0, len(kv)), make([]string, 0, len(kv))
	for _, b := range kv {
		keys, buckets = append(keys, b.Key), append(buckets, b.Bucket)
	}
	if err := sharedName(keys, buckets); err != nil {
		return err
	}

	objects := n.AllObjectStoreBuckets()
	keys, buckets = make([]string, 0, len(objects)), make([]string, 0, len(objects))
	for _, b := range objects {
		keys, buckets = append(keys, b.Key), append(buckets, b.Bucket)
	}

	return sharedName(keys, buckets)
}

// sharedName returns an error for the first bucket that two keys name.
func sharedName(
	keys []string,
	buckets []string,
) error {
	seen := map[string]string{}
	for i, bucket := range buckets {
		if bucket == "" {
			continue
		}
		if key, ok := seen[bucket]; ok {
			return fmt.Errorf(
				"%s and %s are both %q: each needs its own bucket",
				key, keys[i], bucket,
			)
		}
		seen[bucket] = keys[i]
	}

	return nil
}
