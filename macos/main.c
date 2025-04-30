// macos_screen_capture.c
// Compile with: gcc -framework CoreGraphics -framework ApplicationServices
// -framework CoreFoundation -lcurl macos_screen_capture.c -o screen_streamer

#include <ApplicationServices/ApplicationServices.h>
#include <CoreGraphics/CoreGraphics.h>
#include <curl/curl.h>
#include <pthread.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <time.h>
#include <unistd.h>

// Configuration
#define SERVER_URL "http://localhost:8080/upload"
#define JPEG_QUALITY 85
#define CAPTURE_FPS 15
#define FRAME_BUFFER_SIZE 3

// Structure to hold capture settings
typedef struct {
  int width;
  int height;
  int fps;
  int quality;
  const char *server_url;
  int running;
  pthread_mutex_t mutex;
  pthread_cond_t cond;
} CaptureConfig;

// Structure for frame buffer management
typedef struct {
  unsigned char *data;
  size_t size;
  int is_filled;
} FrameBuffer;

// Global variables
static CaptureConfig config;
static FrameBuffer frame_buffers[FRAME_BUFFER_SIZE];
static int current_producer_buffer = 0;
static int current_consumer_buffer = 0;

// Forward declarations
void *capture_thread(void *arg);
void *network_thread(void *arg);
unsigned char *create_jpeg_from_screenshot(CGImageRef image, int quality,
                                           size_t *out_size);

// Initialize the screen capture system
int initialize_capture_system() {
  // Initialize mutex and condition variable
  pthread_mutex_init(&config.mutex, NULL);
  pthread_cond_init(&config.cond, NULL);

  // Get the main display
  CGDirectDisplayID display_id = CGMainDisplayID();

  // Get dimensions of the display
  CGRect bounds = CGDisplayBounds(display_id);
  config.width = bounds.size.width;
  config.height = bounds.size.height;
  config.fps = CAPTURE_FPS;
  config.quality = JPEG_QUALITY;
  config.server_url = SERVER_URL;
  config.running = 1;

  printf("Screen dimensions: %dx%d\n", config.width, config.height);

  // Initialize frame buffers
  for (int i = 0; i < FRAME_BUFFER_SIZE; i++) {
    // Allocate enough memory for a raw screenshot (4 bytes per pixel: RGBA)
    // This is an overestimation since JPEG will be smaller, but ensures we have
    // enough space
    frame_buffers[i].data =
        (unsigned char *)malloc(config.width * config.height * 4);
    frame_buffers[i].size = 0;
    frame_buffers[i].is_filled = 0;

    if (!frame_buffers[i].data) {
      fprintf(stderr, "Failed to allocate memory for frame buffer %d\n", i);
      return 0;
    }
  }

  return 1;
}

// Clean up resources
void cleanup() {
  // Free frame buffers
  for (int i = 0; i < FRAME_BUFFER_SIZE; i++) {
    if (frame_buffers[i].data) {
      free(frame_buffers[i].data);
      frame_buffers[i].data = NULL;
    }
  }

  // Destroy mutex and condition variable
  pthread_mutex_destroy(&config.mutex);
  pthread_cond_destroy(&config.cond);
}

// Main function
int main() {
  printf("Starting macOS Screen Capture for Golang HTTP Server\n");

  // Initialize libcurl
  curl_global_init(CURL_GLOBAL_ALL);

  // Initialize the capture system
  if (!initialize_capture_system()) {
    fprintf(stderr, "Failed to initialize capture system\n");
    return 1;
  }

  // Create threads
  pthread_t capture_thread_id, network_thread_id;
  pthread_create(&capture_thread_id, NULL, capture_thread, NULL);
  pthread_create(&network_thread_id, NULL, network_thread, NULL);

  printf("Press Enter to stop capture...\n");
  getchar();

  // Signal threads to stop
  pthread_mutex_lock(&config.mutex);
  config.running = 0;
  pthread_cond_signal(&config.cond);
  pthread_mutex_unlock(&config.mutex);

  // Wait for threads to finish
  pthread_join(capture_thread_id, NULL);
  pthread_join(network_thread_id, NULL);

  // Clean up
  cleanup();
  curl_global_cleanup();

  printf("Screen capture stopped.\n");
  return 0;
}

// Thread function for screen capture
void *capture_thread(void *arg) {
  printf("Capture thread started\n");

  // Calculate time between frames
  long frame_interval_us = 1000000 / config.fps;
  struct timespec sleep_time;

  while (config.running) {
    // Record start time
    struct timespec start_time;
    clock_gettime(CLOCK_MONOTONIC, &start_time);

    // Get main display ID
    CGDirectDisplayID display_id = CGMainDisplayID();

    // Create a CGImage from the display
    CGImageRef screenshot = CGDisplayCreateImage(display_id);

    if (screenshot) {
      // Convert to JPEG and get the data
      size_t jpeg_size = 0;
      unsigned char *jpeg_data =
          create_jpeg_from_screenshot(screenshot, config.quality, &jpeg_size);

      if (jpeg_data) {
        // Lock mutex before accessing shared data
        pthread_mutex_lock(&config.mutex);

        // Wait if the current buffer is still being processed
        while (config.running &&
               frame_buffers[current_producer_buffer].is_filled) {
          pthread_cond_wait(&config.cond, &config.mutex);
        }

        // Check if we should still be running
        if (!config.running) {
          pthread_mutex_unlock(&config.mutex);
          free(jpeg_data);
          CGImageRelease(screenshot);
          break;
        }

        // Copy JPEG data to the current buffer
        memcpy(frame_buffers[current_producer_buffer].data, jpeg_data,
               jpeg_size);
        frame_buffers[current_producer_buffer].size = jpeg_size;
        frame_buffers[current_producer_buffer].is_filled = 1;

        // Advance to the next producer buffer
        current_producer_buffer =
            (current_producer_buffer + 1) % FRAME_BUFFER_SIZE;

        // Signal network thread that a new frame is available
        pthread_cond_signal(&config.cond);
        pthread_mutex_unlock(&config.mutex);

        // Free the JPEG data
        free(jpeg_data);
      }

      // Release the CGImage
      CGImageRelease(screenshot);
    }

    // Calculate time spent and sleep if needed
    struct timespec end_time;
    clock_gettime(CLOCK_MONOTONIC, &end_time);

    long elapsed_us = (end_time.tv_sec - start_time.tv_sec) * 1000000 +
                      (end_time.tv_nsec - start_time.tv_nsec) / 1000;

    if (elapsed_us < frame_interval_us) {
      long sleep_us = frame_interval_us - elapsed_us;
      sleep_time.tv_sec = sleep_us / 1000000;
      sleep_time.tv_nsec = (sleep_us % 1000000) * 1000;
      nanosleep(&sleep_time, NULL);
    }
  }

  printf("Capture thread stopped\n");
  return NULL;
}

// Convert CGImage to JPEG
unsigned char *create_jpeg_from_screenshot(CGImageRef image, int quality,
                                           size_t *out_size) {
  // We'll use a CFData to store the JPEG data
  CFMutableDataRef jpeg_data = CFDataCreateMutable(kCFAllocatorDefault, 0);
  if (!jpeg_data)
    return NULL;

  // Create a CGImageDestination for JPEG output
  CGImageDestinationRef destination =
      CGImageDestinationCreateWithData(jpeg_data, kUTTypeJPEG, 1, NULL);
  if (!destination) {
    CFRelease(jpeg_data);
    return NULL;
  }

  // Set image properties - JPEG quality
  CFMutableDictionaryRef properties = CFDictionaryCreateMutable(
      kCFAllocatorDefault, 0, &kCFTypeDictionaryKeyCallBacks,
      &kCFTypeDictionaryValueCallBacks);
  CFNumberRef quality_value =
      CFNumberCreate(kCFAllocatorDefault, kCFNumberIntType, &quality);
  CFDictionarySetValue(properties, kCGImageDestinationLossyCompressionQuality,
                       quality_value);
  CGImageDestinationSetProperties(destination, properties);

  // Add the image to the destination
  CGImageDestinationAddImage(destination, image, properties);

  // Finalize the destination to create the JPEG data
  bool success = CGImageDestinationFinalize(destination);

  // Clean up CF objects
  CFRelease(quality_value);
  CFRelease(properties);
  CFRelease(destination);

  if (!success) {
    CFRelease(jpeg_data);
    return NULL;
  }

  // Get the size and data from the CFData
  *out_size = CFDataGetLength(jpeg_data);
  unsigned char *result = (unsigned char *)malloc(*out_size);
  if (result) {
    memcpy(result, CFDataGetBytePtr(jpeg_data), *out_size);
  }

  CFRelease(jpeg_data);
  return result;
}

// Thread function for network communication
void *network_thread(void *arg) {
  printf("Network thread started\n");

  CURL *curl = curl_easy_init();
  if (!curl) {
    fprintf(stderr, "Failed to initialize curl\n");
    return NULL;
  }

  // Set up curl options
  curl_easy_setopt(curl, CURLOPT_URL, config.server_url);
  curl_easy_setopt(curl, CURLOPT_TIMEOUT, 10);
  curl_easy_setopt(curl, CURLOPT_POST, 1L);

  while (config.running) {
    // Lock mutex before accessing shared data
    pthread_mutex_lock(&config.mutex);

    // Wait if no frames are available
    while (config.running &&
           !frame_buffers[current_consumer_buffer].is_filled) {
      pthread_cond_wait(&config.cond, &config.mutex);
    }

    // Check if we should still be running
    if (!config.running) {
      pthread_mutex_unlock(&config.mutex);
      break;
    }

    // Get the size and pointer to the current buffer
    size_t frame_size = frame_buffers[current_consumer_buffer].size;
    unsigned char *frame_data = frame_buffers[current_consumer_buffer].data;

    // Mark the buffer as not filled
    frame_buffers[current_consumer_buffer].is_filled = 0;

    // Advance to the next consumer buffer
    int buffer_to_process = current_consumer_buffer;
    current_consumer_buffer = (current_consumer_buffer + 1) % FRAME_BUFFER_SIZE;

    // Signal that we've freed up a buffer
    pthread_cond_signal(&config.cond);
    pthread_mutex_unlock(&config.mutex);

    // Send the frame to the server
    struct curl_slist *headers = NULL;
    headers = curl_slist_append(headers, "Content-Type: image/jpeg");

    curl_easy_setopt(curl, CURLOPT_HTTPHEADER, headers);
    curl_easy_setopt(curl, CURLOPT_POSTFIELDSIZE, frame_size);
    curl_easy_setopt(curl, CURLOPT_POSTFIELDS, frame_data);

    CURLcode res = curl_easy_perform(curl);

    curl_slist_free_all(headers);

    if (res != CURLE_OK) {
      fprintf(stderr, "Failed to send frame: %s\n", curl_easy_strerror(res));
      usleep(100000); // Sleep a bit before retrying
    }
  }

  curl_easy_cleanup(curl);
  printf("Network thread stopped\n");
  return NULL;
}
