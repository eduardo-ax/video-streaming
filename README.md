# Video Store API

**Video Store** is a RESTful microservice that handles video uploads, metadata storage, and cloud-based delivery through **HLS streaming**.
It integrates with **Kafka** for event-driven processing and stores both metadata and files in scalable cloud services.

---

## Endpoints

### `POST v1/videos`

Uploads a new video and stores its metadata.

#### **Request**

**Content-Type:** `multipart/form-data`
Fields:

* `title` — video title
* `description` — video description
* `file` — video file (.mp4, .mov, etc.)

#### **Example**

```bash
curl -X POST http://localhost:8080/v1/videos \
  -F "title=My Test Video" \
  -F "description=First upload using HLS" \
  -F "file=@/path/to/video.mp4"
```

#### **Internal process**

1. The service generates a unique `id` using a **BIGSERIAL** primary key from the `videos` table.
2. The `title` and `description` are saved in the relational database.
3. The video file is uploaded to the **S3 bucket**.
4. A **Kafka** message is published with:

   * **key:** `id`
   * **value:** `id/filename`
     (example: `42/video123.mp4`)
5. Another service (e.g., a transcoder) can then convert this file into **HLS segments** (`.m3u8` and `.ts` files).

---

### `GET /videos/v1/:id/:filename`

Streams the video using **HTTP Live Streaming (HLS)** format.

This endpoint delivers:

* `.m3u8` playlist files — which describe available video segments
* `.ts` chunk files — which contain the actual video segments

#### **Example**

When integrated with a player (like **HTML5**, **Video.js**, or **hls.js**):

```html
<video controls autoplay width="640" height="360">
  <source src="http://localhost:8080/videos/42/index.m3u8" type="application/x-mpegURL">
</video>
```

The player will automatically request the `.m3u8` playlist and sequential `.ts` chunks for playback.

---

## Database Structure

**Table:** `videos`

| Column      | Type         | Description               |
| ----------- | ------------ | ------------------------- |
| id          | BIGSERIAL PK | Unique video identifier   |
| title       | TEXT         | Video title               |
| description | TEXT         | Video description         |

---

##  Kafka Integration

After a successful upload, a message is sent to the configured Kafka topic in the following format:

```json
{
  "key": "42",
  "value": "42/video123.mp4"
}
```

This allows other microservices (e.g., transcoding or CDN distribution) to process the video asynchronously.

---

## Cloud Storage

Uploaded files are stored in an S3-compatible bucket.
The storage path follows this structure:

```
videos/{id}/{filename}
```

Example:

```
videos/42/video123.mp4
```


# Video Transcoding Service

**Video Transcoding Service** is a worker that listens to a **Kafka** topic, downloads raw video files from cloud storage (S3), transcodes them into **HLS format** using **FFmpeg**, and uploads the generated `.m3u8` playlist and `.ts` chunks back to the same S3 path.
After processing, it removes all temporary local files to keep the container clean.

---

## Overview

This service works together with the [Video Store API](../video-store-api) as part of a cloud-based video pipeline.

### Workflow

1. **Video Store API** publishes a message to Kafka when a new video is uploaded:

   ```json
   {
     "key": "42",
     "value": "42/video123.mp4"
   }
   ```
2. **Video Transcoding Service** consumes this message from the Kafka topic `transcoding`.
3. It downloads the source file (`video123.mp4`) from S3 using the path `videos/42/video123.mp4`.
4. The service runs **FFmpeg** to generate **HLS output**:

   * `index.m3u8`
   * `index0.ts`, `index1.ts`, `index2.ts`, ...
5. The `.m3u8` and `.ts` files are uploaded back to the **same S3 folder**:

   ```
   videos/42/
     ├── index.m3u8
     ├── index0.ts
     ├── index1.ts
     ├── ...
   ```
6. Local temporary files are deleted after upload to save container space.

---

## Kafka

* **Topic:** `transcoding`
* **Key:** video ID (e.g., `42`)
* **Value:** path to the uploaded file (e.g., `42/video123.mp4`)

Each message triggers one transcoding job.

---

## FFmpeg Command

The service uses **FFmpeg** inside the container to perform the conversion.

Typical command:

```bash
ffmpeg -i input.mp4 \
  -profile:v baseline -level 3.0 \
  -start_number 0 \
  -hls_time 6 \
  -hls_list_size 0 \
  -f hls index.m3u8
```

This generates:

* **6-second chunks** (`index0.ts`, `index1.ts`, …)
* **One playlist file** (`index.m3u8`)

All output files are stored temporarily in the container before being uploaded back to the S3 bucket.

---

## Cloud Storage (S3)

* **Input file:**
  `videos/{id}/{filename}`
  Example: `videos/42/video123.mp4`

* **Output files (HLS):**

  ```
  videos/{id}/index.m3u8
  videos/{id}/index0.ts
  videos/{id}/index1.ts
  ...
  ```

---

## Cleanup

After successfully uploading the `.m3u8` and `.ts` files to S3, the service deletes:

* The original downloaded video file
* All generated local HLS chunks

This ensures that the container filesystem stays lightweight and avoids unnecessary disk usage.

---

## Technologies

* **Go (Golang)** — backend language
* **Echo** — HTTP framework
* **PostgreSQL** — relational database
* **Kafka** — message broker
* **AWS S3 (or compatible)** — cloud storage
* **HLS (HTTP Live Streaming)** — for adaptive video delivery

---

## Environment Variables

| Variable                | Description                                           |
| ----------------------- | ----------------------------------------------------- |
| `KAFKA_BROKER`          | Kafka broker address                                  |
| `KAFKA_TOPIC`           | Topic to listen for messages (default: `transcoding`) |
| `S3_BUCKET`             | Target S3 bucket name                                 |
| `S3_REGION`             | AWS region or compatible endpoint                     |
| `AWS_ACCESS_KEY_ID`     | AWS access key                                        |
| `AWS_SECRET_ACCESS_KEY` | AWS secret key                                        |
| `TMP_PATH`              | Local path for temporary files                        |

Example `.env`:

```env
KAFKA_BROKER=localhost:9092
KAFKA_TOPIC=transcoding
S3_BUCKET=video-storage
S3_REGION=us-east-1
AWS_ACCESS_KEY_ID=your-key
AWS_SECRET_ACCESS_KEY=your-secret
TMP_PATH=/tmp/videos
```

---

## Running the Services

```bash
go run main.go
```

Or using Docker Compose:

```bash
docker-compose up --build
```

---

## Author

Developed by **Eduardo** — a cloud-based video storage and streaming API built for scalability and event-driven workflows.
