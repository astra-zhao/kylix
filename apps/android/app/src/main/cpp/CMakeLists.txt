cmake_minimum_required(VERSION 3.22.1)
project(kylixjni)

# libkylixlogic.so is produced by apps/android/build_core.sh and dropped in
# src/main/jniLibs/<abi>/. This JNI library is what Kotlin loads second.
set(KYLIX_LOGIC "${CMAKE_SOURCE_DIR}/../jniLibs/${ANDROID_ABI}/libkylixlogic.so")
if(NOT EXISTS "${KYLIX_LOGIC}")
    message(FATAL_ERROR
        "Missing ${KYLIX_LOGIC}. From the repo root run: bash apps/android/build_core.sh ${ANDROID_ABI}")
endif()

add_library(kylixlogic SHARED IMPORTED)
set_target_properties(kylixlogic PROPERTIES IMPORTED_LOCATION "${KYLIX_LOGIC}")

add_library(kylixjni SHARED kylix_jni.c)
target_link_libraries(kylixjni kylixlogic)
