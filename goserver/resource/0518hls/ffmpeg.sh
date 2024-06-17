ffmpeg -i ../0518sample.mp4 -codec: copy -start_number 0 -hls_time 1 -hls_list_size 0 -f hls 0518sample.m3u8
