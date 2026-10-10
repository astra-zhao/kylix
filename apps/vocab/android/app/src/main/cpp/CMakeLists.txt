cmake_minimum_required(VERSION 3.22.1)
project(vocabjni)

# libkylixvocab.so is produced by apps/vocab/android/build_core.sh and dropped
# in src/main/jniLibs/<abi>/. This JNI library is what Kotlin loads second.
set(KYLIX_LOGIC "${CMAKE_SOURCE_DIR}/../jniLibs/${ANDROID_ABI}/libkylixvocab.so")
if(NOT EXISTS "${KYLIX_LOGIC}")
    message(FATAL_ERROR
        "Missing ${KYLIX_LOGIC}. From the repo root run: bash apps/vocab/android/build_core.sh ${ANDROID_ABI}")
endif()

add_library(kylixvocab SHARED IMPORTED)
set_target_properties(kylixvocab PROPERTIES IMPORTED_LOCATION "${KYLIX_LOGIC}")

add_library(vocabjni SHARED kylix_jni.c)
target_link_libraries(vocabjni kylixvocab)
